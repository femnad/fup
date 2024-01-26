package provision

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gabriel-vasile/mimetype"
	"github.com/ulikunitz/xz"

	"github.com/femnad/fup/common"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/precheck/when"
	"github.com/femnad/fup/remote"
	"github.com/femnad/fup/settings"
	"github.com/femnad/mare"
	marecmd "github.com/femnad/mare/cmd"
)

const (
	bufferSize         = 8192
	bzipMimeType       = "application/x-bzip2"
	dirMode            = 0755
	executableMimeType = "application/x-executable"
	githubReleaseRegex = "https://github.com/[a-zA-Z_-]+/[a-zA-Z_-]+/releases/download/[v0-9.]+/[a-zA-Z0-9_.-]+"
	gzipMimeType       = "application/gzip"
	sharedLibMimeType  = "application/x-sharedlib"
	tarMimeType        = "application/x-tar"
	xzMimeType         = "application/x-xz"
	zipMimeType        = "application/zip"
)

type archiveEntry struct {
	info os.FileInfo
	name string
}

type extractionHint struct {
	file     string
	fileType string
	target   string
}

type extractionFn func(entity.Release, extractionHint) (ReleaseInfo, error)

// ReleaseInfo stores an archive's root dir and specifies if the root dir is part of the archive files.
type ReleaseInfo struct {
	execCandidate  string
	hasRootDir     bool
	target         string
	targetOverride string
}

func (r ReleaseInfo) GetTarget() string {
	if r.targetOverride != "" {
		return r.targetOverride
	}

	return r.target
}

func downloadRelease(release entity.Release, s settings.Settings) (string, error) {
	releaseURL, err := release.ExpandURL(s)
	if err != nil {
		return "", err
	}

	if releaseURL == "" {
		return "", fmt.Errorf("no URL given for release %v", release)
	}
	internal.Log.Infof("Downloading %s", releaseURL)

	response, err := remote.ReadResponseBody(releaseURL)
	if err != nil {
		return "", err
	}

	return downloadTempFile(response)
}

func processDownload(release entity.Release, s settings.Settings) (info ReleaseInfo, err error) {
	tempFile, err := downloadRelease(release, s)
	if err != nil {
		return
	}

	fileType, err := mimetype.DetectFile(tempFile)
	if err != nil {
		return
	}

	dirName := internal.ExpandUser(s.ReleaseDir)
	err = os.MkdirAll(dirName, dirMode)
	if err != nil {
		return
	}

	extractFn, err := getExtractionFn(fileType.String())
	if err != nil {
		return
	}
	info, err = extractFn(release, extractionHint{
		file:     tempFile,
		fileType: fileType.String(),
		target:   dirName,
	})
	if err != nil {
		return
	}

	info.target, err = getAbsTarget(dirName, info)
	if err != nil {
		return
	}

	err = os.Remove(tempFile)
	if err != nil {
		return info, err
	}

	return
}

func getTarReader(reader io.ReadCloser, fileType string) (io.Reader, error) {
	switch fileType {
	case gzipMimeType:
		return gzip.NewReader(reader)
	case bzipMimeType:
		return bzip2.NewReader(reader), nil
	case tarMimeType:
		return reader, nil
	case xzMimeType:
		return xz.NewReader(reader)
	default:
		return nil, fmt.Errorf("unable to determine tar reader for file type %s", fileType)
	}
}

func extractCompressedFile(info os.FileInfo, outputPath string, reader io.Reader) error {
	if info.IsDir() {
		if err := os.MkdirAll(outputPath, info.Mode()); err != nil {
			return err
		}
		return nil
	}

	dir := path.Dir(outputPath)
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return err
	}

	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
	if err != nil {
		return err
	}

	_, err = io.Copy(file, reader)
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}

func unzipFile(f *zip.File, outputPath string) error {
	info := f.FileInfo()
	if info.IsDir() {
		if err := os.MkdirAll(outputPath, info.Mode()); err != nil {
			return err
		}
		return nil
	}

	dir := path.Dir(outputPath)
	if err := os.MkdirAll(dir, dirMode); err != nil {
		return err
	}

	file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
	if err != nil {
		return err
	}

	fileInArchive, err := f.Open()
	if err != nil {
		return err
	}

	_, err = io.Copy(file, fileInArchive)
	if err != nil {
		return err
	}

	err = file.Close()
	if err != nil {
		return err
	}

	err = fileInArchive.Close()
	if err != nil {
		return err
	}

	return nil
}

func downloadTempFile(response remote.Response) (string, error) {
	tempFile, err := os.CreateTemp("/tmp", "*")
	if err != nil {
		return "", err
	}
	err = tempFile.Close()
	if err != nil {
		return "", err
	}

	tempFilePath := tempFile.Name()

	err = download(response.Body, tempFilePath)
	if err != nil {
		return "", err
	}

	return tempFilePath, nil
}

func commonPrefix(names []string) string {
	if len(names) == 0 {
		return ""
	}

	first := names[0]
	minLength := len(first)
	for _, name := range names {
		if len(name) < minLength {
			minLength = len(name)
		}
	}

	for i := 0; i < minLength; i++ {
		currentChar := first[i]
		for _, name := range names {
			if name[i] != currentChar {
				return first[:i]
			}
		}
	}

	return first[:minLength]
}

func getExecCandidate(prefix, execCandidate string) (string, error) {
	if prefix == "" {
		return execCandidate, nil
	}

	if !strings.Contains(prefix, "/") {
		return execCandidate, nil
	}

	if execCandidate == prefix {
		tokens := strings.SplitN(execCandidate, "/", 2)
		if len(tokens) != 2 {
			return "", fmt.Errorf("error determining executable candidate with prefix %s", prefix)
		}
		return tokens[1], nil
	}

	return strings.TrimPrefix(execCandidate, prefix), nil
}

func getReleaseInfo(archive entity.Release, entries []archiveEntry) (info ReleaseInfo, err error) {
	names := mare.Map(entries, func(entry archiveEntry) string {
		return entry.name
	})
	prefix := commonPrefix(names)
	roots := mapset.NewSet[string]()

	var execs []archiveEntry
	for _, entry := range entries {
		rootDir := strings.Split(entry.name, "/")
		roots.Add(rootDir[0])
		if common.IsExecutableFile(entry.info) {
			execs = append(execs, entry)
		}
	}

	var execCandidate string
	var hasRootDir bool
	var target string

	if roots.Cardinality() == 1 {
		root, ok := roots.Pop()
		if !ok {
			return info, fmt.Errorf("error determining root dir for %s", archive.Url)
		}

		hasRootDir = strings.Index(prefix, "/") > -1
		target = root
	} else if archive.Name() != "" {
		target = archive.Name()
	} else {
		target = execs[0].name
	}

	if len(execs) == 1 {
		execCandidate = execs[0].name
	}

	execCandidate, err = getExecCandidate(prefix, execCandidate)
	if err != nil {
		return
	}

	return ReleaseInfo{
		execCandidate:  execCandidate,
		hasRootDir:     hasRootDir,
		target:         target,
		targetOverride: archive.Target}, nil
}

func getOutputPath(info ReleaseInfo, fileName, dirName string) string {
	if info.hasRootDir {
		if info.targetOverride != "" && strings.HasPrefix(fileName, info.target) {
			fileName = strings.Replace(fileName, info.target, info.targetOverride, 1)
		}
		return filepath.Join(dirName, fileName)
	}

	return filepath.Join(dirName, info.GetTarget(), fileName)
}

func getAbsTarget(dirName string, info ReleaseInfo) (string, error) {
	if path.IsAbs(info.target) {
		return info.target, nil
	}

	if path.IsAbs(dirName) {
		return path.Join(dirName, info.target), nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return path.Join(wd, dirName, info.GetTarget()), nil
}

func getTarEntries(tempFile, fileType string) (entries []archiveEntry, err error) {
	f, err := os.Open(tempFile)
	if err != nil {
		return
	}
	defer f.Close()

	reader, err := getTarReader(f, fileType)
	if err != nil {
		return
	}

	var header *tar.Header
	tarReader := tar.NewReader(reader)
	for {
		header, err = tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return
		}

		entries = append(entries, archiveEntry{
			info: header.FileInfo(),
			name: header.Name,
		})
	}

	return entries, nil
}

// Shamelessly lifted from https://golangdocs.com/tar-gzip-in-golang
func untar(archive entity.Release, source extractionHint) (info ReleaseInfo, err error) {
	entries, err := getTarEntries(source.file, source.fileType)
	if err != nil {
		return
	}

	info, err = getReleaseInfo(archive, entries)
	if err != nil {
		return
	}

	f, err := os.Open(source.file)
	if err != nil {
		return
	}
	defer f.Close()

	reader, err := getTarReader(f, source.fileType)
	if err != nil {
		return
	}

	var header *tar.Header
	tarReader := tar.NewReader(reader)
	for {
		header, err = tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return
		}

		outputPath := getOutputPath(info, header.Name, source.target)
		err = extractCompressedFile(header.FileInfo(), outputPath, tarReader)
		if err != nil {
			return
		}
	}

	return info, nil
}

func getZipInfo(tempFile string) (entries []archiveEntry, err error) {
	zipArchive, err := zip.OpenReader(tempFile)
	if err != nil {
		return
	}

	for _, f := range zipArchive.File {
		entries = append(entries, archiveEntry{
			info: f.FileInfo(),
			name: f.Name,
		})
	}

	err = zipArchive.Close()
	if err != nil {
		return
	}

	return entries, nil
}

func unzip(archive entity.Release, source extractionHint) (info ReleaseInfo, err error) {
	entries, err := getZipInfo(source.file)
	if err != nil {
		return
	}

	info, err = getReleaseInfo(archive, entries)
	if err != nil {
		return
	}

	zipArchive, err := zip.OpenReader(source.file)
	if err != nil {
		return
	}

	for _, f := range zipArchive.File {
		output := getOutputPath(info, f.Name, source.target)
		err = unzipFile(f, output)
		if err != nil {
			return
		}
	}

	return info, nil
}

func copyBinary(release entity.Release, hint extractionHint) (info ReleaseInfo, err error) {
	name := release.Name()
	if name == "" {
		_, name = path.Split(release.Url)
	}
	target := release.Target
	if target == "" {
		target = name
	}

	src, err := os.Open(hint.file)
	if err != nil {
		return
	}

	copyTarget := path.Join(hint.target, target, name)
	copyTargetDir, _ := path.Split(copyTarget)
	err = ensureDirExist(copyTargetDir)
	if err != nil {
		return
	}

	dst, err := os.OpenFile(copyTarget, os.O_CREATE|os.O_WRONLY, os.FileMode(0o755))
	if err != nil {
		return
	}

	_, err = io.Copy(dst, src)
	if err != nil {
		return
	}

	err = dst.Close()
	if err != nil {
		return
	}

	return ReleaseInfo{execCandidate: name, hasRootDir: true, target: target}, nil
}

func getExtractionFn(fileType string) (extractionFn, error) {
	switch fileType {
	case executableMimeType, sharedLibMimeType:
		return copyBinary, nil
	case bzipMimeType, gzipMimeType, tarMimeType, xzMimeType:
		return untar, nil
	case zipMimeType:
		return unzip, nil
	default:
		return nil, fmt.Errorf("unable to determine extractor for file type %s", fileType)
	}
}

func Extract(archive entity.Release, s settings.Settings) (ReleaseInfo, error) {
	return processDownload(archive, s)
}

func download(closer io.ReadCloser, target string) error {
	f, err := os.Create(target)
	if err != nil {
		return fmt.Errorf("error creating target %s: %w", target, err)
	}

	for {
		buf := make([]byte, bufferSize)

		readBytes, readErr := closer.Read(buf)
		if !errors.Is(readErr, io.EOF) && err != nil {
			return readErr
		}

		_, writeErr := f.Write(buf[:readBytes])
		if writeErr != nil {
			return err
		}
		if errors.Is(readErr, io.EOF) {
			break
		}
	}

	err = closer.Close()
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return nil

}

func guessArchiveName(releaseUrl string) (string, error) {
	pattern := regexp.MustCompile(githubReleaseRegex)
	if !pattern.MatchString(releaseUrl) {
		return "", nil
	}

	parsed, err := url.Parse(releaseUrl)
	if err != nil {
		return "", err
	}

	return strings.Split(parsed.Path, "/")[2], nil
}

func ensureRelease(release entity.Release, s settings.Settings) error {
	releaseURL, err := release.ExpandURL(s)
	if err != nil {
		return err
	}

	if !when.ShouldRun(release) {
		internal.Log.Debugf("Skipping extracting release %s due to when condition %s", releaseURL, release.When)
		return nil
	}

	if release.Name() == "" {
		name, err := guessArchiveName(releaseURL)
		if err != nil {
			return err
		}
		release.Ref = name
	}

	if unless.ShouldSkip(release, s) {
		internal.Log.Debugf("Skipping download: %s", releaseURL)
		return nil
	}

	info, err := Extract(release, s)
	if err != nil {
		internal.Log.Errorf("Error downloading release %s: %v", releaseURL, err)
		return err
	}

	target := info.target
	if info.targetOverride != "" {
		target, _ = path.Split(target)
		target = path.Join(target, info.targetOverride)
	}
	for _, symlink := range release.ExpandSymlinks(info.execCandidate) {
		err = createSymlink(symlink, target, s.GetBinPath())
		if err != nil {
			internal.Log.Errorf("error creating symlink for release %s: %v", releaseURL, err)
			return err
		}
	}

	version := release.Version
	if version == "" {
		version = s.Versions[release.Name()]
	}

	executeAfter := release.ExecuteAfter
	for _, cmd := range executeAfter.Cmd {
		cmd = settings.ExpandStringWithLookup(s, cmd, map[string]string{"version": version})
		pwd := ""
		if executeAfter.SetPwd {
			pwd = info.GetTarget()
		}
		if pwd == "" {
			internal.Log.Debugf("Running command %s", cmd)
		} else {
			internal.Log.Debugf("Running command %s under path %s", cmd, pwd)
		}

		_, err = marecmd.RunFormatError(marecmd.Input{Command: cmd, Pwd: pwd, Shell: true})
		if err != nil {
			internal.Log.Errorf("error running post download command: %v", err)
			return err
		}
	}

	return nil
}

func ensureReleases(releases []entity.Release, s settings.Settings) error {
	var releaseErrs []error
	for _, archive := range releases {
		err := ensureRelease(archive, s)
		releaseErrs = append(releaseErrs, err)
	}

	return errors.Join(releaseErrs...)
}

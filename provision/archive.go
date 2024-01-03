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
	"github.com/xi2/xz"

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
	githubReleaseRegex = "https://github.com/[a-zA-Z_-]+/[a-zA-Z_-]+/releases/download/[v0-9.]+/[a-zA-Z0-9_.-]+"
	gzipMimeType       = "application/gzip"
	tarMimeType        = "application/x-tar"
	xzDictMax          = 1 << 27
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

type extractionFn func(entity.Archive, extractionHint) (ArchiveInfo, error)

// ArchiveInfo stores an archive's root dir and species if the root dir is part of the archive files.
type ArchiveInfo struct {
	hasRootDir     bool
	maybeExec      string
	target         string
	targetOverride string
}

func (a ArchiveInfo) GetTarget() string {
	if a.targetOverride != "" {
		return a.targetOverride
	}

	return a.target
}

func downloadRelease(archive entity.Archive, s settings.Settings) (string, error) {
	archiveURL, err := archive.ExpandURL(s)
	if err != nil {
		return "", err
	}

	if archiveURL == "" {
		return "", fmt.Errorf("no URL given for archive %v", archive)
	}
	internal.Log.Infof("Downloading %s", archiveURL)

	response, err := remote.ReadResponseBody(archiveURL)
	if err != nil {
		return "", err
	}

	return downloadTempFile(response)
}

func processDownload(archive entity.Archive, s settings.Settings) (info ArchiveInfo, err error) {
	tempFile, err := downloadRelease(archive, s)
	if err != nil {
		return
	}

	fileType, err := mimetype.DetectFile(tempFile)
	if err != nil {
		return
	}

	extractFn, err := getExtractionFn(fileType.String())
	if err != nil {
		return
	}

	dirName := internal.ExpandUser(s.ExtractDir)
	err = os.MkdirAll(dirName, dirMode)
	if err != nil {
		return
	}

	info, err = extractFn(archive, extractionHint{
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
		return xz.NewReader(reader, xzDictMax)
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

func getArchiveInfo(archive entity.Archive, entries []archiveEntry) (ArchiveInfo, error) {
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

	var maybeExec string
	var hasRootDir bool
	var target string

	if roots.Cardinality() == 1 {
		root, ok := roots.Pop()
		if !ok {
			return ArchiveInfo{}, fmt.Errorf("error determining root dir for %s", archive.Url)
		}

		hasRootDir = strings.Index(prefix, "/") > -1
		target = root
	} else if archive.Name() != "" {
		target = archive.Name()
	} else {
		target = execs[0].name
	}

	if len(execs) == 1 {
		maybeExec = execs[0].name
	}

	return ArchiveInfo{hasRootDir: hasRootDir, maybeExec: maybeExec, target: target, targetOverride: archive.Target}, nil
}

func getOutputPath(info ArchiveInfo, fileName, dirName string) string {
	if info.hasRootDir {
		if info.targetOverride != "" && strings.HasPrefix(fileName, info.target) {
			fileName = strings.Replace(fileName, info.target, info.targetOverride, 1)
		}
		return filepath.Join(dirName, fileName)
	}

	return filepath.Join(dirName, info.GetTarget(), fileName)
}

func getAbsTarget(dirName string, info ArchiveInfo) (string, error) {
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
func untar(archive entity.Archive, source extractionHint) (info ArchiveInfo, err error) {
	entries, err := getTarEntries(source.file, source.fileType)
	if err != nil {
		return
	}

	info, err = getArchiveInfo(archive, entries)
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

func unzip(archive entity.Archive, source extractionHint) (info ArchiveInfo, err error) {
	entries, err := getZipInfo(source.file)
	if err != nil {
		return
	}

	info, err = getArchiveInfo(archive, entries)
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

func getExtractionFn(fileType string) (extractionFn, error) {
	switch fileType {
	case bzipMimeType, gzipMimeType, tarMimeType, xzMimeType:
		return untar, nil
	case zipMimeType:
		return unzip, nil
	default:
		return nil, fmt.Errorf("unable to determine extractor for file type %s", fileType)
	}
}

func Extract(archive entity.Archive, s settings.Settings) (ArchiveInfo, error) {
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

func extractArchive(archive entity.Archive, s settings.Settings) error {
	archiveUrl, err := archive.ExpandURL(s)
	if err != nil {
		return err
	}

	if !when.ShouldRun(archive) {
		internal.Log.Debugf("Skipping extracting archive %s due to when condition %s", archiveUrl, archive.When)
		return nil
	}

	if archive.Name() == "" {
		name, err := guessArchiveName(archiveUrl)
		if err != nil {
			return err
		}
		archive.Ref = name
	}

	if unless.ShouldSkip(archive, s) {
		internal.Log.Debugf("Skipping download: %s", archiveUrl)
		return nil
	}

	info, err := Extract(archive, s)
	if err != nil {
		internal.Log.Errorf("Error downloading archive %s: %v", archiveUrl, err)
		return err
	}

	target := info.target
	if info.targetOverride != "" {
		target, _ = path.Split(target)
		target = path.Join(target, info.targetOverride)
	}
	for _, symlink := range archive.ExpandSymlinks(info.maybeExec) {
		err = createSymlink(symlink, target, s.GetBinPath())
		if err != nil {
			internal.Log.Errorf("error creating symlink for archive %s: %v", archiveUrl, err)
			return err
		}
	}

	version := archive.Version
	if version == "" {
		version = s.Versions[archive.Name()]
	}

	for _, cmd := range archive.ExecuteAfter {
		cmd = settings.ExpandStringWithLookup(s, cmd, map[string]string{"version": version})
		internal.Log.Debugf("Running command %s", cmd)
		_, err = marecmd.RunFormatError(marecmd.Input{Command: cmd, Shell: true})
		if err != nil {
			internal.Log.Errorf("error running post extract command: %v", err)
			return err
		}
	}

	return nil
}

func extractArchives(archives []entity.Archive, s settings.Settings) error {
	var archiveErrs []error
	for _, archive := range archives {
		err := extractArchive(archive, s)
		archiveErrs = append(archiveErrs, err)
	}

	return errors.Join(archiveErrs...)
}

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
	rootUser           = "root"
	setuidExecutable   = 0o4755
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

type extractionFn func(ReleaseInfo, extractionHint) error

// ReleaseInfo stores an archive's root dir and specifies if the root dir is part of the archive files.
type ReleaseInfo struct {
	execCandidate  string
	hasRootDir     bool
	absTarget      string
	relTarget      string
	targetOverride string
}

type executionCtx struct {
	s             settings.Settings
	releaseTarget string
	version       string
}

func (r ReleaseInfo) GetTarget() string {
	if r.targetOverride != "" {
		return r.targetOverride
	}

	return r.relTarget
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
	hint := extractionHint{
		file:     tempFile,
		fileType: fileType.String(),
		target:   dirName,
	}
	info, err = getInfo(release, hint)
	if err != nil {
		return info, err
	}

	absTarget, err := getAbsTarget(dirName, info)
	if err != nil {
		return
	}

	if release.Cleanup {
		internal.Log.Debugf("Purging directory %s", absTarget)
		err = os.RemoveAll(absTarget)
		if err != nil {
			return info, fmt.Errorf("error cleaning up dir %s before extraction: %v", absTarget, err)
		}
	}

	var chromeSandbox string
	if release.ChromeSandbox != "" {
		chromeSandbox = path.Join(absTarget, release.ChromeSandbox)
		err = internal.EnsureFileAbsent(chromeSandbox)
		if err != nil {
			return info, fmt.Errorf("error removing chrome-sandbox file %s: %v", chromeSandbox, err)
		}
	}

	extractFn, err := getExtractionFn(fileType.String())
	if err != nil {
		return
	}
	err = extractFn(info, hint)
	if err != nil {
		return
	}

	if chromeSandbox != "" {
		err = internal.Chown(chromeSandbox, rootUser, rootUser)
		if err != nil {
			return info, err
		}

		err = internal.Chmod(chromeSandbox, setuidExecutable)
		if err != nil {
			return info, err
		}
	}

	err = os.Remove(tempFile)
	if err != nil {
		return
	}

	info.absTarget = absTarget
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

func getReleaseInfo(release entity.Release, entries []archiveEntry) (info ReleaseInfo, err error) {
	names := mare.Map(entries, func(entry archiveEntry) string {
		return entry.name
	})
	prefix := commonPrefix(names)
	roots := mapset.NewSet[string]()

	var execs []archiveEntry
	for _, entry := range entries {
		name := strings.TrimPrefix(entry.name, "./")
		if name == "" {
			continue
		}

		rootDir := strings.Split(name, "/")
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
			return info, fmt.Errorf("error determining root dir for %s", release.Url)
		}

		prefix = strings.TrimPrefix(prefix, "./")
		hasRootDir = strings.Index(prefix, "/") > -1
		if root == "." {
			target = prefix
		} else {
			target = root
		}
	} else if release.Name() != "" {
		target = release.Name()
	} else {
		target = execs[0].name
	}

	numExecs := len(execs)
	if numExecs == 1 {
		execCandidate = strings.TrimPrefix(execs[0].name, "./")
	} else if numExecs > 1 {
		for _, exec := range execs {
			dir, baseName := path.Split(exec.name)
			if baseName == release.Ref {
				execCandidate = baseName
				target = strings.TrimSuffix(strings.TrimPrefix(dir, "./"), "/")
				break
			}
		}
	}

	execCandidate, err = getExecCandidate(prefix, execCandidate)
	if err != nil {
		return
	}

	return ReleaseInfo{
		execCandidate:  execCandidate,
		hasRootDir:     hasRootDir,
		relTarget:      target,
		targetOverride: release.Target}, nil
}

func getOutputPath(info ReleaseInfo, fileName, dirName string) string {
	if info.hasRootDir {
		if info.targetOverride != "" && strings.HasPrefix(fileName, info.relTarget) {
			fileName = strings.Replace(fileName, info.relTarget, info.targetOverride, 1)
		}
		return filepath.Join(dirName, fileName)
	}

	return filepath.Join(dirName, info.GetTarget(), fileName)
}

func getAbsTarget(dirName string, info ReleaseInfo) (string, error) {
	if path.IsAbs(info.relTarget) {
		return info.relTarget, nil
	}

	if path.IsAbs(dirName) {
		return path.Join(dirName, info.relTarget), nil
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

func untarInfo(release entity.Release, hint extractionHint) (info ReleaseInfo, err error) {
	entries, err := getTarEntries(hint.file, hint.fileType)
	if err != nil {
		return
	}

	return getReleaseInfo(release, entries)
}

// Shamelessly lifted from https://golangdocs.com/tar-gzip-in-golang
func untar(info ReleaseInfo, source extractionHint) error {
	f, err := os.Open(source.file)
	if err != nil {
		return err
	}
	defer f.Close()

	reader, err := getTarReader(f, source.fileType)
	if err != nil {
		return err
	}

	var header *tar.Header
	tarReader := tar.NewReader(reader)
	for {
		header, err = tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		outputPath := getOutputPath(info, header.Name, source.target)
		err = extractCompressedFile(header.FileInfo(), outputPath, tarReader)
		if err != nil {
			return err
		}
	}

	return nil
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

func unzipInfo(release entity.Release, hint extractionHint) (info ReleaseInfo, err error) {
	entries, err := getZipInfo(hint.file)
	if err != nil {
		return
	}

	return getReleaseInfo(release, entries)
}

func unzip(info ReleaseInfo, source extractionHint) (err error) {
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

	return nil
}

func binaryInfo(release entity.Release, _ extractionHint) (info ReleaseInfo, err error) {
	name := release.Name()
	if name == "" {
		_, name = path.Split(release.Url)
	}
	target := release.Target
	if target == "" {
		target = name
	}

	return ReleaseInfo{execCandidate: name, hasRootDir: true, relTarget: target}, nil
}

func copyBinary(info ReleaseInfo, hint extractionHint) (err error) {
	src, err := os.Open(hint.file)
	if err != nil {
		return
	}

	copyTarget := path.Join(hint.target, info.relTarget, info.execCandidate)
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

	return dst.Close()
}

func getInfo(release entity.Release, hint extractionHint) (ReleaseInfo, error) {
	switch hint.fileType {
	case executableMimeType, sharedLibMimeType:
		return binaryInfo(release, hint)
	case bzipMimeType, gzipMimeType, tarMimeType, xzMimeType:
		return untarInfo(release, hint)
	case zipMimeType:
		return unzipInfo(release, hint)
	default:
		return ReleaseInfo{}, fmt.Errorf("unable to determine extractor for file type %s", hint.fileType)
	}
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
		if readErr != nil && !errors.Is(readErr, io.EOF) {
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

func performExecutions(eCtx executionCtx, spec entity.ExecuteSpec) error {
	for _, cmd := range spec.Cmd {
		cmd = settings.ExpandStringWithLookup(eCtx.s, cmd, map[string]string{"version": eCtx.version})
		pwd := ""
		if spec.SetPwd {
			pwd = path.Join(eCtx.s.ReleaseDir, eCtx.releaseTarget)
			pwd = internal.ExpandUser(pwd)
		}
		if pwd == "" {
			internal.Log.Debugf("Running command %s", cmd)
		} else {
			internal.Log.Debugf("Running command %s under path %s", cmd, pwd)
		}

		err := marecmd.RunErrOnly(marecmd.Input{Command: cmd, Pwd: pwd, Shell: true, Sudo: spec.Sudo})
		if err != nil {
			internal.Log.Errorf("error running post download command: %v", err)
			return err
		}
	}

	return nil
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
		var name string
		name, err = guessArchiveName(releaseURL)
		if err != nil {
			return err
		}
		release.Ref = name
	}

	if unless.ShouldSkip(release, s) {
		internal.Log.Debugf("Skipping download: %s", releaseURL)
		return nil
	}

	version := release.Version
	if version == "" {
		version = s.Versions[release.Name()]
	}

	eCtx := executionCtx{s: s, version: version}
	err = performExecutions(eCtx, release.ExecuteBefore)
	if err != nil {
		return err
	}

	info, err := Extract(release, s)
	if err != nil {
		internal.Log.Errorf("Error downloading release %s: %v", releaseURL, err)
		return err
	}

	target := info.absTarget
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

	eCtx.releaseTarget = info.GetTarget()
	return performExecutions(eCtx, release.ExecuteAfter)
}

func processGithubReleases(githubReleases []entity.GithubRelease) ([]entity.Release, error) {
	var releases []entity.Release
	for _, githubRelease := range githubReleases {
		if githubRelease.Ref == "" {
			return releases, fmt.Errorf("no ref specified for GitHub release: %+v", githubRelease)
		}

		githubRef := githubRelease.Ref
		releaseUrl := fmt.Sprintf("https://github.com/%s/releases/download/%s", githubRef,
			githubRelease.Url)

		ref := githubRelease.ExecName
		if ref == "" {
			refTokens := strings.Split(githubRef, "/")
			if len(refTokens) != 2 {
				return releases, fmt.Errorf("unexpected release name %s", githubRef)
			}
			ref = refTokens[1]
		}

		release := entity.Release{
			Cleanup:       githubRelease.Cleanup,
			DontLink:      githubRelease.DontLink,
			DontUpdate:    githubRelease.DontUpdate,
			ExecuteAfter:  githubRelease.ExecuteAfter,
			ExecuteBefore: githubRelease.ExecuteBefore,
			NamedLink:     githubRelease.NamedLink,
			Ref:           ref,
			Symlink:       githubRelease.Symlink,
			Target:        githubRelease.Target,
			Unless:        githubRelease.Unless,
			Url:           releaseUrl,
			Version:       githubRelease.Version,
			VersionLookup: githubRelease.VersionLookup,
			When:          githubRelease.When,
		}
		releases = append(releases, release)
	}

	return releases, nil
}

func ghCliAvailable(s settings.Settings) bool {
	if !s.UseGHClient {
		return false
	}

	err := marecmd.RunErrOnly(marecmd.Input{Command: "gh auth status"})
	return err == nil
}

func ensureReleases(config entity.Config) error {
	var releaseErrs []error
	config.Settings.Internal = settings.InternalSettings{GhAvailable: ghCliAvailable(config.Settings)}

	releases := config.Releases
	processedReleases, err := processGithubReleases(config.GithubReleases)
	if err == nil {
		releases = append(releases, processedReleases...)
	} else {
		releaseErrs = append(releaseErrs, err)
	}

	for _, release := range releases {
		err = ensureRelease(release, config.Settings)
		if err != nil {
			internal.Log.Errorf("Error ensuring release %s: %v", release.Name(), err)
		}
		releaseErrs = append(releaseErrs, err)
	}

	return errors.Join(releaseErrs...)
}

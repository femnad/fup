package provision

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/xi2/xz"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/base/settings"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/precheck/unless"
	"github.com/femnad/fup/precheck/when"
	"github.com/femnad/fup/remote"
	"github.com/femnad/mare"
	marecmd "github.com/femnad/mare/cmd"
)

const (
	bufferSize   = 8192
	dirMode      = 0755
	xzDictMax    = 1 << 27
	tarFileRegex = `\.tar(\.(gz|bz2|xz))?$`
)

type archiveEntry struct {
	info os.FileInfo
	name string
}

// archiveInfo stores an archive's root dir and species if the root dir is part of the archive files.
type archiveInfo struct {
	hasRootDir bool
	maybeExec  string
	target     string
}

func processDownload(archive base.Archive, s settings.Settings) (archiveInfo, error) {
	var info archiveInfo
	url := archive.ExpandURL(s)
	if url == "" {
		return info, fmt.Errorf("no URL given for archive %v", archive)
	}
	internal.Log.Infof("Downloading %s", url)

	response, err := remote.ReadResponseBody(url)
	if err != nil {
		return info, err
	}

	extractFn, err := getExtractionFn(archive, s, response.ContentDisposition)
	if err != nil {
		return info, err
	}

	dirName := internal.ExpandUser(s.ExtractDir)
	err = os.MkdirAll(dirName, dirMode)
	if err != nil {
		return info, err
	}

	return extractFn(archive, response, dirName)
}

func getReader(response remote.Response, tempFile *os.File) (io.Reader, error) {
	filename := getFilename(response)
	if strings.HasSuffix(filename, ".tar") {
		return tempFile, nil
	}

	if strings.HasSuffix(filename, ".tar.gz") {
		gzipReader, err := gzip.NewReader(tempFile)
		if err != nil {
			return nil, err
		}
		return gzipReader, nil
	}

	if strings.HasSuffix(filename, ".tar.bz2") {
		return bzip2.NewReader(tempFile), nil
	}

	if strings.HasSuffix(filename, ".tar.xz") {
		xzReader, err := xz.NewReader(tempFile, xzDictMax)
		if err != nil {
			return nil, err
		}
		return xzReader, nil
	}

	return nil, fmt.Errorf("unable to determine tar reader for URL %s", response.URL)
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

func isExecutableFile(info os.FileInfo) bool {
	return !info.IsDir() && info.Mode().Perm()&0100 != 0
}

func getFilename(response remote.Response) string {
	filename := response.URL
	if response.ContentDisposition != "" {
		filename = response.ContentDisposition
	}
	return filename
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

func getArchiveInfo(archive base.Archive, entries []archiveEntry) (archiveInfo, error) {
	names := mare.Map(entries, func(entry archiveEntry) string {
		return entry.name
	})
	prefix := commonPrefix(names)
	roots := mapset.NewSet[string]()
	var execs []archiveEntry
	for _, entry := range entries {
		rootDir := strings.Split(entry.name, "/")
		roots.Add(rootDir[0])
		if isExecutableFile(entry.info) {
			execs = append(execs, entry)
		}
	}

	var maybeExec string
	var hasRootDir bool
	var target string
	if roots.Cardinality() == 1 {
		root, ok := roots.Pop()
		if !ok {
			return archiveInfo{}, fmt.Errorf("error determining root dir for %s", archive.Url)
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

	return archiveInfo{hasRootDir: hasRootDir, maybeExec: maybeExec, target: target}, nil
}

func getTarInfo(archive base.Archive, response remote.Response, tempfile string) (archiveInfo, error) {
	f, err := os.Open(tempfile)
	if err != nil {
		return archiveInfo{}, err
	}
	defer f.Close()

	reader, err := getReader(response, f)
	if err != nil {
		return archiveInfo{}, err
	}

	var entries []archiveEntry
	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			panic(err)
		}

		entries = append(entries, archiveEntry{
			info: header.FileInfo(),
			name: header.Name,
		})
	}

	return getArchiveInfo(archive, entries)
}

func getOutputPath(info archiveInfo, fileName, dirName string) string {
	if info.hasRootDir {
		return filepath.Join(dirName, fileName)
	}

	return filepath.Join(dirName, info.target, fileName)
}

func getAbsTarget(dirName string, info archiveInfo) (string, error) {
	if path.IsAbs(dirName) {
		return path.Join(dirName, info.target), nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return path.Join(wd, dirName, info.target), nil
}

// Shamelessly lifted from https://golangdocs.com/tar-gzip-in-golang
func untar(archive base.Archive, response remote.Response, dirName string) (archiveInfo, error) {
	var info archiveInfo
	tempfile, err := downloadTempFile(response)
	if err != nil {
		return info, err
	}

	info, err = getTarInfo(archive, response, tempfile)

	f, err := os.Open(tempfile)
	if err != nil {
		return info, err
	}
	defer f.Close()

	reader, err := getReader(response, f)
	if err != nil {
		return info, err
	}

	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			panic(err)
		}

		outputPath := getOutputPath(info, header.Name, dirName)
		err = extractCompressedFile(header.FileInfo(), outputPath, tarReader)
		if err != nil {
			return info, err
		}
	}

	err = os.Remove(tempfile)
	if err != nil {
		return info, err
	}

	info.target, err = getAbsTarget(dirName, info)
	if err != nil {
		return info, err
	}

	return info, nil
}

func getZipInfo(archive base.Archive, tempFile string) (archiveInfo, error) {
	var info archiveInfo
	zipArchive, err := zip.OpenReader(tempFile)
	if err != nil {
		return info, err
	}

	var entries []archiveEntry
	for _, f := range zipArchive.File {
		entries = append(entries, archiveEntry{
			info: f.FileInfo(),
			name: f.Name,
		})
	}

	err = zipArchive.Close()
	if err != nil {
		return info, err
	}

	return getArchiveInfo(archive, entries)
}

func unzip(archive base.Archive, response remote.Response, dirName string) (archiveInfo, error) {
	var info archiveInfo
	tempFile, err := downloadTempFile(response)

	root, err := getZipInfo(archive, tempFile)
	if err != nil {
		return info, err
	}

	zipArchive, err := zip.OpenReader(tempFile)
	if err != nil {
		return info, err
	}

	for _, f := range zipArchive.File {
		output := getOutputPath(root, f.Name, dirName)
		err = unzipFile(f, output)
		if err != nil {
			return info, err
		}
	}

	err = os.Remove(tempFile)
	if err != nil {
		return info, err
	}

	info.target, err = getAbsTarget(dirName, root)
	if err != nil {
		return info, err
	}

	return info, nil
}

func getExtractionFn(archive base.Archive, s settings.Settings, contentDisposition string) (func(base.Archive, remote.Response, string) (archiveInfo, error), error) {
	fileName := archive.ExpandURL(s)
	if contentDisposition != "" {
		fileName = contentDisposition
	}

	tarRegex := regexp.MustCompile(tarFileRegex)
	if tarRegex.MatchString(fileName) {
		return untar, nil
	}

	if strings.HasSuffix(fileName, ".zip") {
		return unzip, nil
	}

	return nil, fmt.Errorf("unable to find extraction method for URL %s", fileName)
}

func Extract(archive base.Archive, s settings.Settings) (archiveInfo, error) {
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

func extractArchive(archive base.Archive, s settings.Settings) error {
	url := archive.ExpandURL(s)

	if !when.ShouldRun(archive) {
		internal.Log.Debugf("Skipping extracting archive %s due to when condition %s", url, archive.When)
		return nil
	}

	if unless.ShouldSkip(archive, s) {
		internal.Log.Debugf("Skipping download: %s", url)
		return nil
	}

	info, err := Extract(archive, s)
	if err != nil {
		internal.Log.Errorf("Error downloading archive %s: %v", url, err)
		return err
	}

	for _, symlink := range archive.ExpandSymlinks(s, info.maybeExec) {
		err = createSymlink(symlink, info.target, s.GetBinPath())
		if err != nil {
			internal.Log.Errorf("error creating symlink for archive %s: %v", url, err)
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

func extractArchives(archives []base.Archive, s settings.Settings) error {
	var archiveErrs []error
	for _, archive := range archives {
		err := extractArchive(archive, s)
		archiveErrs = append(archiveErrs, err)
	}

	return errors.Join(archiveErrs...)
}

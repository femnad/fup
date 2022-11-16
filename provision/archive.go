package provision

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/xi2/xz"

	"github.com/femnad/fup/base"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/remote"
)

const (
	bufferSize   = 8192
	dirMode      = 0755
	tarFileRegex = `\.tar(\.(gz|bz2|xz))$`
)

func processDownload(archive base.Archive, archiveDir string, processor func(io.ReadCloser, base.Archive, string) error) error {
	url := archive.ExpandURL()
	if url == "" {
		return fmt.Errorf("no URL given for archive %v", archive)
	}
	internal.Log.Infof("Downloading %s", url)

	respBody, err := remote.ReadResponseBody(url)
	if err != nil {
		return err
	}

	dirName := internal.ExpandUser(archiveDir)
	err = os.MkdirAll(dirName, dirMode)
	if err != nil {
		return err
	}

	if archive.Binary != "" {
		dirName = filepath.Join(dirName, archive.Binary)
	}

	return processor(respBody, archive, dirName)
}

func mkdirAll(dir string, mode os.FileMode) error {
	err := os.MkdirAll(dir, mode)
	if err != nil {
		return err
	}

	return nil
}

func getReader(tarfile io.ReadCloser, archive base.Archive) (io.Reader, error) {
	url := archive.ExpandURL()

	if strings.HasSuffix(url, ".tar") {
		return tarfile, nil
	}

	if strings.HasSuffix(url, ".tar.gz") {
		gzipReader, err := gzip.NewReader(tarfile)
		if err != nil {
			return nil, err
		}
		return gzipReader, nil
	}

	if strings.HasSuffix(url, ".tar.bz2") {
		return bzip2.NewReader(tarfile), nil
	}

	if strings.HasSuffix(url, ".tar.xz") {
		xzReader, err := xz.NewReader(tarfile, 0)
		if err != nil {
			return nil, err
		}
		return xzReader, nil
	}

	return nil, fmt.Errorf("unable to determine tar reader for URL %s", url)
}

func extractCompressedFile(info os.FileInfo, outputPath string, reader io.Reader) error {
	if info.IsDir() {
		if err := mkdirAll(outputPath, info.Mode()); err != nil {
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

func unzipFile(info os.FileInfo, outputPath string, f *zip.File) error {
	if info.IsDir() {
		if err := mkdirAll(outputPath, info.Mode()); err != nil {
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

// Shamelessly lifted from https://golangdocs.com/tar-gzip-in-golang
func untar(tarfile io.ReadCloser, archive base.Archive, target string) error {
	defer func() {
		err := tarfile.Close()
		if err != nil {
			log.Fatalf("Error closing tarfile: %v", err)
		}
	}()

	reader, err := getReader(tarfile, archive)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(reader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			panic(err)
		}

		outputPath := filepath.Join(target, header.Name)
		info := header.FileInfo()
		err = extractCompressedFile(info, outputPath, tarReader)
		if err != nil {
			return err
		}
	}

	return nil
}

func unzip(zipfile io.ReadCloser, archive base.Archive, target string) error {
	tempFile, err := os.CreateTemp("/tmp", "*.zip")
	if err != nil {
		return err
	}
	err = tempFile.Close()
	if err != nil {
		return err
	}

	tempFilePath := tempFile.Name()
	err = download(zipfile, archive, tempFilePath)
	if err != nil {
		return err
	}

	zipArchive, err := zip.OpenReader(tempFilePath)
	if err != nil {
		return err
	}

	for _, f := range zipArchive.File {
		fp := filepath.Join(target, f.Name)
		err := unzipFile(f.FileInfo(), fp, f)
		if err != nil {
			return err
		}
	}

	err = os.Remove(tempFilePath)
	if err != nil {
		return err
	}

	return nil
}

func getExtractionFn(archive base.Archive) (func(io.ReadCloser, base.Archive, string) error, error) {
	url := archive.ExpandURL()
	tarRegex := regexp.MustCompile(tarFileRegex)
	if tarRegex.MatchString(url) {
		return untar, nil
	}

	if strings.HasSuffix(url, ".zip") {
		return unzip, nil
	}

	return nil, fmt.Errorf("unable find extraction method for URL %s", url)
}

func Extract(archive base.Archive, archiveDir string) error {
	extractFn, err := getExtractionFn(archive)
	if err != nil {
		return err
	}

	return processDownload(archive, archiveDir, extractFn)
}

func download(closer io.ReadCloser, archive base.Archive, target string) error {
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

func Download(archive base.Archive, archiveDir string) error {
	return processDownload(archive, archiveDir, download)
}

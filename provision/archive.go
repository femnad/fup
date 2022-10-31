package provision

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"github.com/femnad/fup/remote"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/femnad/fup/base"
)

const (
	bufferSize = 8192
	dirMode    = 0755
)

func expandUser(path string) string {
	return strings.Replace(path, "~", os.Getenv("HOME"), 1)
}

func processDownload(archive base.Archive, archiveDir string, processor func(closer io.ReadCloser, target string) error) error {
	url := archive.Url
	if url == "" {
		return fmt.Errorf("no URL given for archive %v", archive)
	}

	url = os.Expand(url, archive.ExpandArchive)
	log.Printf("Downloading %s", url)

	respBody, err := remote.ReadResponseBody(url)
	if err != nil {
		return err
	}

	dirName := expandUser(archiveDir)
	err = os.MkdirAll(dirName, dirMode)
	if err != nil {
		return err
	}

	if archive.Binary != "" {
		dirName = filepath.Join(dirName, archive.Binary)
	}

	return processor(respBody, dirName)
}

func mkdirAll(dir string, mode os.FileMode) error {
	err := os.MkdirAll(dir, mode)
	if err != nil {
		return err
	}

	return nil
}

// Shamelessly lifted from https://golangdocs.com/tar-gzip-in-golang
func untar(tarfile io.ReadCloser, target string) error {
	defer func() {
		err := tarfile.Close()
		if err != nil {
			log.Fatalf("Error closing tarfile: %v", err)
		}
	}()

	gzipReader, err := gzip.NewReader(tarfile)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			panic(err)
		}

		outputPath := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = mkdirAll(outputPath, info.Mode()); err != nil {
				return err
			}
			continue
		}

		dir, _ := path.Split(outputPath)
		if err = os.MkdirAll(dir, dirMode); err != nil {
			return err
		}
		file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}

		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}

		err = file.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func Extract(archive base.Archive, archiveDir string) error {
	return processDownload(archive, archiveDir, untar)
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

func Download(archive base.Archive, archiveDir string) error {
	return processDownload(archive, archiveDir, download)
}

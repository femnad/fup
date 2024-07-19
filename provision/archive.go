package provision

import (
	"archive/tar"
	"errors"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/femnad/fup/entity"
	"github.com/femnad/fup/internal"
	"github.com/femnad/fup/remote"
	"io"
	"os"
	"path"
	"regexp"
	"strings"
)

var (
	mimeTypeMap = map[string]string{
		"bz2": bzipMimeType,
		"gz":  gzipMimeType,
		"xz":  xzMimeType,
		"tar": tarMimeType,
	}
	tarFileRegex = regexp.MustCompile(".*\\.tar(\\.(bz2|gz|xz))?")
)

func extractTar(reader *tar.Reader, outputPath string, fileSet mapset.Set[string]) error {
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return err
		}

		info := header.FileInfo()
		name := header.Name
		if !fileSet.Contains(name) {
			continue
		}

		target := path.Join(outputPath, name)
		err = extractCompressedFile(info, target, reader)
		if err != nil {
			return err
		}
	}

	return nil
}

func getMimeType(filename string) (string, error) {
	dotIndex := strings.LastIndex(filename, ".")
	if dotIndex < 0 {
		return "", fmt.Errorf("unable to determine extension for filename %s", filename)
	}

	extension := filename[dotIndex+1:]
	mime, ok := mimeTypeMap[extension]
	if !ok {
		return "", fmt.Errorf("unable to determine mime type for filename %s", filename)
	}

	return mime, nil
}

func extract(response remote.Response, archive entity.Archive) error {
	archiveUrl := response.URL
	disposition := response.ContentDisposition
	outputPath := internal.ExpandUser(archive.Target)

	fileSet := mapset.NewSet[string]()
	for _, file := range archive.Files {
		fileSet.Add(file)
	}

	if tarFileRegex.MatchString(disposition) {
		mime, err := getMimeType(disposition)
		if err != nil {
			return err
		}

		reader, err := getTarReader(response.Body, mime)
		if err != nil {
			return err
		}

		return extractTar(tar.NewReader(reader), outputPath, fileSet)
	}

	return fmt.Errorf("unable to determine archive reader for %s", archiveUrl)
}

func shouldSkip(archive entity.Archive) (bool, error) {
	for _, file := range archive.Files {
		target := internal.ExpandUser(path.Join(archive.Target, file))
		_, err := os.Stat(target)
		if err == nil {
			continue
		}
		if !os.IsNotExist(err) {
			return false, err
		} else {
			return false, nil
		}
	}

	return true, nil
}

func extractArchive(archive entity.Archive) error {
	skip, err := shouldSkip(archive)
	if err != nil {
		return err
	}

	archiveURL := archive.URL
	if skip {
		internal.Log.Debugf("Skipping extracting archive %s", archiveURL)
		return nil
	}

	internal.Log.Infof("Extracting archive %s", archiveURL)

	response, err := remote.ReadResponseBody(archiveURL)
	if err != nil {
		return err
	}

	return extract(response, archive)
}

func extractArchives(config entity.Config) error {
	var errs []error
	for _, archive := range config.Archives {
		err := extractArchive(archive)
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

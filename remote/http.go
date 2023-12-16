package remote

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/femnad/fup/internal"
)

const (
	userAgentKey = "user-agent"
	userAgent    = "femnad/fup"
	utfPrefix    = "UTF-8''"
)

var (
	okStatuses = []int{http.StatusOK}
)

type Response struct {
	Body               io.ReadCloser
	ContentDisposition string
	URL                string
}

func getAttachmentFilename(header http.Header) string {
	contentDispositionValue := header.Get("Content-Disposition")
	for _, value := range strings.Split(contentDispositionValue, "; ") {
		if value == "attachment" {
			continue
		}

		fields := strings.SplitN(value, "=", 2)
		if len(fields) != 2 {
			continue
		}

		filename := fields[1]
		if strings.HasPrefix(filename, utfPrefix) {
			return strings.TrimPrefix(filename, utfPrefix)
		}

		return filename
	}

	return ""
}

func ReadResponseBody(url string) (Response, error) {
	var response Response
	cl := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return response, err
	}
	req.Header.Set(userAgentKey, userAgent)

	resp, err := cl.Do(req)
	if err != nil {
		return response, err
	}

	statusCode := resp.StatusCode
	if !internal.Contains(okStatuses, statusCode) {
		return response, fmt.Errorf("error reading response, got status %d from URL %s", statusCode, url)
	}

	attachmentFilename := getAttachmentFilename(resp.Header)
	response = Response{Body: resp.Body, ContentDisposition: attachmentFilename, URL: url}

	return response, nil
}

func ReadResponseBytes(url string) ([]byte, error) {
	response, err := ReadResponseBody(url)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(response.Body)
}

func Download(url, target string) error {
	if url == "" {
		return fmt.Errorf("download URL is empty")
	}
	if target == "" {
		return fmt.Errorf("download target is empty")
	}

	resp, err := ReadResponseBody(url)
	if err != nil {
		return err
	}

	dir, _ := path.Split(target)
	if err = internal.EnsureDir(dir); err != nil {
		return err
	}

	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(0o644))
	if err != nil {
		return err
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

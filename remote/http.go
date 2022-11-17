package remote

import (
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/femnad/fup/internal"
)

const (
	contentDispositionPattern = `attachment; filename="(.*)"`
	userAgentKey              = "user-agent"
	userAgent                 = "femnad/fup"
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
	matches := regexp.MustCompile(contentDispositionPattern).FindStringSubmatch(contentDispositionValue)
	if len(matches) != 2 {
		return ""
	}

	return matches[1]
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

package remote

import (
	"fmt"
	"github.com/femnad/fup/internal"
	"io"
	"net/http"
)

const (
	userAgentKey = "user-agent"
	userAgent    = "femnad/fup"
)

var (
	okStatuses = []int{http.StatusOK}
)

func ReadResponseBody(url string) (io.ReadCloser, error) {
	cl := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(userAgentKey, userAgent)

	resp, err := cl.Do(req)
	if err != nil {
		return nil, err
	}

	statusCode := resp.StatusCode
	if !internal.Contains(okStatuses, statusCode) {
		return nil, fmt.Errorf("error reading response, got status %d from URL %s", statusCode, url)
	}

	return resp.Body, nil
}

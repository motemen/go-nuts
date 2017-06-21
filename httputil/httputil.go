package httputil

import (
	"net/http"
)

type HTTPError struct {
	StatusCode int
	Status     string
}

func (e *HTTPError) Error() string {
	return e.Status
}

func Successful(resp *http.Response, err error) (*http.Response, error) {
	if err != nil {
		return resp, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp, &HTTPError{StatusCode: resp.StatusCode, Status: resp.Status}
	}
	return resp, nil
}

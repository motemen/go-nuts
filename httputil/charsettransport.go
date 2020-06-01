package httputil

import (
	"bytes"
	"io"
	"net/http"

	"golang.org/x/net/html/charset"
)

// CharsetTransport is an http.Transport which automatically decodes resp.Body by its charset.
type CharsetTransport struct {
	Base http.RoundTripper
}

type readCloser struct {
	io.Reader
	io.Closer
}

// RoundTrip implements http.RoundTripper.
func (t *CharsetTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	resp, err := base.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	r, err := charset.NewReader(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil && err != io.EOF {
		return resp, err
	}

	if r == nil {
		r = bytes.NewReader(nil)
	}

	resp.Body = &readCloser{
		Reader: r,
		Closer: resp.Body,
	}
	return resp, nil
}

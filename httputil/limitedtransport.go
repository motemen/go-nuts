package httputil

import (
	"io"
	"net/http"
)

type LimitedTransport struct {
	Base http.RoundTripper
	N    int64
}

func (t *LimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	resp, err := base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	resp.Body = readCloser{
		Reader: io.LimitReader(resp.Body, t.N),
		Closer: resp.Body,
	}
	return resp, nil
}

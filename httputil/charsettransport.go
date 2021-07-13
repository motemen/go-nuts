package httputil

import (
	"bytes"
	"io"
	"net/http"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"

	"github.com/motemen/go-nuts/chardet"
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

type ChardetTransport struct {
	Base          http.RoundTripper
	DetectOptions []chardet.DetectOption
}

func (t *ChardetTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}

	resp, err := base.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// from golang.org/x/net/html/charset.NewReader
	var r io.Reader

	preview := make([]byte, 1024)
	n, err := io.ReadFull(resp.Body, preview)
	switch {
	case err == io.ErrUnexpectedEOF:
		preview = preview[:n]
		r = bytes.NewReader(preview)
	case err == io.EOF:
		r = bytes.NewReader(nil)
	case err != nil:
		return nil, err
	default:
		r = io.MultiReader(bytes.NewReader(preview), r)
	}

	if n > 0 {
		enc, _, certain := charset.DetermineEncoding(preview, resp.Header.Get("Content-Type"))
		if !certain {
			if e, _ := chardet.DetectEncoding(preview, t.DetectOptions...); e != nil {
				enc = e
			}
		}

		if enc != nil && enc != encoding.Nop {
			r = transform.NewReader(r, enc.NewDecoder())
		}
	}

	resp.Body = &readCloser{
		Reader: r,
		Closer: resp.Body,
	}
	return resp, nil
}

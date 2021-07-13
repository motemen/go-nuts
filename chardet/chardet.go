package chardet

import (
	"sort"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"

	"github.com/saintfish/chardet"
)

var detector = chardet.NewTextDetector()

type DetectOption func(detectOption) detectOption

type detectOption struct {
	charsets  []string
	languages []string
	prefer    func(a, b chardet.Result) bool // return true if a is preferable than b
}

var ErrNotDetected = chardet.NotDetectedError

func (dc detectOption) filter(results []chardet.Result) []chardet.Result {
	if dc.charsets != nil {
		m := map[string]bool{}
		for _, k := range dc.charsets {
			m[k] = true
		}

		filtered := []chardet.Result{}
		for _, r := range results {
			if m[r.Charset] {
				filtered = append(filtered, r)
			}
		}

		results = filtered
	}

	if dc.languages != nil {
		m := map[string]bool{}
		for _, k := range dc.languages {
			m[k] = true
		}

		filtered := []chardet.Result{}
		for _, r := range results {
			if m[r.Language] {
				filtered = append(filtered, r)
			}
		}

		results = filtered
	}

	if dc.prefer != nil {
		sort.Slice(
			results,
			func(i, j int) bool {
				return dc.prefer(results[i], results[j])
			},
		)
	}

	return results
}

func DetectEncoding(b []byte, opts ...DetectOption) (encoding.Encoding, error) {
	results, err := detector.DetectAll(b)
	if err != nil {
		return nil, err
	}

	var dc detectOption
	for _, o := range opts {
		dc = o(dc)
	}

	results = dc.filter(results)
	if len(results) == 0 {
		return nil, ErrNotDetected
	}

	return ianaindex.IANA.Encoding(results[0].Charset)
}

func WithCharset(charsets ...string) DetectOption {
	return func(dc detectOption) detectOption {
		dc.charsets = charsets
		return dc
	}
}

func WithLanguage(langs ...string) DetectOption {
	return func(dc detectOption) detectOption {
		dc.languages = langs
		return dc
	}
}

func WithPrefer(f func(a, b chardet.Result) bool) DetectOption {
	return func(dc detectOption) detectOption {
		dc.prefer = f
		return dc
	}
}

package chardet

import (
	"sort"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"

	"github.com/saintfish/chardet"
)

var detector = chardet.NewTextDetector()

type Detector struct {
	resultFilter
}

type DetectorOption func(resultFilter) resultFilter

var ErrNotDetected = chardet.NotDetectedError

func NewDetector(opts ...DetectorOption) *Detector {
	var d Detector
	for _, o := range opts {
		d.resultFilter = o(d.resultFilter)
	}
	return &d
}

func (d Detector) DetectEncoding(b []byte) (encoding.Encoding, string) {
	results, err := detector.DetectAll(b)
	if err != nil {
		return nil, ""
	}

	results = d.resultFilter.filter(results)
	if len(results) == 0 {
		return nil, ""
	}

	charset := results[0].Charset

	enc, err := ianaindex.IANA.Encoding(charset)
	if err != nil {
		return nil, ""
	}

	return enc, charset
}

func WithCharset(charsets ...string) DetectorOption {
	return func(dc resultFilter) resultFilter {
		dc.charsets = charsets
		return dc
	}
}

func WithLanguage(langs ...string) DetectorOption {
	return func(dc resultFilter) resultFilter {
		dc.languages = langs
		return dc
	}
}

func WithPrefer(f func(a, b chardet.Result) bool) DetectorOption {
	return func(dc resultFilter) resultFilter {
		dc.prefer = f
		return dc
	}
}

type resultFilter struct {
	charsets  []string
	languages []string
	prefer    func(a, b chardet.Result) bool // return true if a is preferable than b
}

func (dc resultFilter) filter(results []chardet.Result) []chardet.Result {
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

package bytesutil

import (
	"strings"

	"golang.org/x/text/transform"
)

type Set interface {
	Contains(b byte) bool
}

type setFunc func(byte) bool

func (s setFunc) Contains(b byte) bool {
	return s(b)
}

func In(s string) Set {
	return setFunc(func(b byte) bool {
		return strings.IndexByte(s, b) != -1
	})
}

func Predicate(f func(byte) bool) Set {
	return setFunc(f)
}

func Remove(s Set) removeTransformer {
	return removeTransformer{s}
}

type removeTransformer struct {
	set Set
}

// Reset implements transform.Transformer.
func (t removeTransformer) Reset() {}

// Transform implements transform.Transformer.
func (t removeTransformer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for nSrc < len(src) {
		p := IndexFunc(src[nSrc:], t.set.Contains)
		if p == 0 {
			nSrc++
			continue
		}

		var end, skip int
		if p == -1 {
			end = len(src)
			skip = 0
		} else {
			end = nSrc + p
			skip = 1
		}

		if cap(dst[nDst:]) < end-nSrc+1 {
			err = transform.ErrShortDst
			return
		}

		n := copy(dst[nDst:], src[nSrc:end])
		nDst += n
		nSrc += n + skip
	}

	return
}

func IndexFunc(s []byte, f func(b byte) bool) int {
	for i := 0; i < len(s); i++ {
		if f(s[i]) {
			return i
		}
	}
	return -1
}

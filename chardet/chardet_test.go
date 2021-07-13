package chardet

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectEncoding(t *testing.T) {
	ff, _ := filepath.Glob("testdata/*.txt")
	for _, f := range ff {
		filename := filepath.Base(f)
		t.Run(filename, func(t *testing.T) {
			b, _ := os.ReadFile(f)
			enc, err := DetectEncoding(b, WithLanguage("ja", ""))
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(
				t,
				strings.ToLower(strings.TrimSuffix(filename, ".txt")),
				strings.Replace(strings.ToLower(fmt.Sprint(enc)), " ", "_", -1),
			)
		})
	}
}

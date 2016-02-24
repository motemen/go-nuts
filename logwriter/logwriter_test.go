package logwriter

import (
	"testing"

	"bytes"
	"fmt"
	"log"
)

func TestLogWriter(t *testing.T) {
	var buf bytes.Buffer
	w := &LogWriter{
		Logger: log.New(&buf, "", log.Lshortfile),
		Format: "[test] %s",
	}

	fmt.Fprintln(w, "foo")
	fmt.Fprint(w, "bar-")

	t.Log(buf.String())

	fmt.Fprintln(w, "baz")

	t.Log(buf.String())
}

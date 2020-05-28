package httputil

import (
	"io/ioutil"
	"mime"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCharsetTransport(t *testing.T) {
	mime.AddExtensionType(".html", "text/html; charset=unknown")

	s := httptest.NewServer(http.FileServer(http.Dir("testdata")))
	defer s.Close()

	client := &http.Client{
		Transport: &CharsetTransport{},
	}

	resp, err := client.Get(s.URL + "/euc-jp.html")
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "こんにちは、世界") {
		t.Fatal(string(b))
	}

	_, err = client.Get(s.URL + "/empty.html")
	if err != nil {
		t.Fatal(err)
	}
}

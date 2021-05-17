package httputil

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestLimitedTransport(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		n, _ := strconv.Atoi(req.URL.Query().Get("n"))
		fmt.Fprint(w, strings.Repeat("x", n))
	})
	s := httptest.NewServer(h)
	defer s.Close()

	client := &http.Client{
		Transport: &LimitedTransport{N: 5000},
	}

	tests := []int{10000, 5000, 300}
	for _, test := range tests {
		resp, err := client.Get(s.URL + "?n=" + fmt.Sprint(test))
		if err != nil {
			t.Fatal(err)
		}

		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}

		expected := test
		if expected > 5000 {
			expected = 5000
		}
		if got := len(b); got != expected {
			t.Errorf("got=%v, expected=%v", got, expected)
		}
	}
}

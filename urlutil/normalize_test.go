package urlutil

import (
	"net/url"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{
			name: "Remove default port (http)",
			in:   "http://www.example.com:80/",
			want: "http://www.example.com/",
		},
		{
			name: "Remove default port (https)",
			in:   "https://www.example.com:443/",
			want: "https://www.example.com/",
		},
		{
			name: "Remove empty port",
			in:   "http://www.example.com:/",
			want: "http://www.example.com/",
		},
		{
			name: "Keep non-default port",
			in:   "http://www.example.com:9999/",
			want: "http://www.example.com:9999/",
		},
		{
			name: "Encode to Punycode",
			in:   "https://はじめよう.みんな/はじめよう.みんな",
			want: "https://xn--p8j9a0d9c9a.xn--q9jyb4c/%E3%81%AF%E3%81%98%E3%82%81%E3%82%88%E3%81%86.%E3%81%BF%E3%82%93%E3%81%AA",
		},
		{
			name: "Encode/decode characters",
			in:   "https://localhost/%7e%41%5E/🤗?q=%7E%41%5E🤗/",
			want: "https://localhost/~A%5E/%F0%9F%A4%97?q=~A%5E%F0%9F%A4%97%2F",
		},
		{
			name: "Encode/decode characters",
			in:   `https://localhost/!"$%ef%41`,
			want: "https://localhost/%21%22$%EFA",
		},
		{
			name: "Uppercase percent encodings",
			in:   "https://localhost/%5e",
			want: "https://localhost/%5E",
		},
		{
			name: "Space in path/query",
			in:   "https://localhost/foo bar?q=foo bar+baz",
			want: "https://localhost/foo%20bar?q=foo+bar+baz",
		},
		{
			name: "Empty path",
			in:   "https://localhost",
			want: "https://localhost/",
		},
		{
			name: "Lowercase scheme/host",
			in:   "HTTPS://WWW.EXAMPLE.COM/",
			want: "https://www.example.com/",
		},
		{
			name: "Clean path",
			in:   "https://localhost/a/./b/../c//d/",
			want: "https://localhost/a/c//d/",
		},
	}
	for _, tt := range tests {
		name := tt.name
		if name == "" {
			name = tt.in
		}
		t.Run(name, func(t *testing.T) {
			u, _ := url.Parse(tt.in)
			got, err := NormalizeURL(u)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.String() != tt.want {
				t.Errorf("NormalizeURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_removeDotSegments(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			in:   "/a/b/c/./../../g",
			want: "/a/g",
		},
		{
			in:   "mid/content=5/../6",
			want: "mid/6",
		},
		{
			in:   "foo/bar/.",
			want: "foo/bar/",
		},
		{
			in:   "foo/bar/..",
			want: "foo/",
		},
		{
			in:   "/../.",
			want: "/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeDotSegments(tt.in); got != tt.want {
				t.Errorf("removeDotSegments(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

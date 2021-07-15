package urlutil

import (
	"net/url"
	"testing"
)

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{
			in:   "http://www.example.com:80/",
			want: "http://www.example.com/",
		},
		{
			in:   "https://www.example.com:443/",
			want: "https://www.example.com/",
		},
		{
			in:   "https://www.example.com:9999/",
			want: "https://www.example.com:9999/",
		},
		{
			in:   "https://はじめよう.みんな/",
			want: "https://xn--p8j9a0d9c9a.xn--q9jyb4c/",
		},
		{
			in:   "https://localhost/abc/🤗",
			want: "https://localhost/abc/%F0%9F%A4%97",
		},
		{
			in:   "https://localhost/?q=🤗",
			want: "https://localhost/?q=%F0%9F%A4%97",
		},
		{
			in:   "https://localhost/?q=%f0%9f%a4%97",
			want: "https://localhost/?q=%F0%9F%A4%97",
		},
		{
			in:   "https://localhost",
			want: "https://localhost/",
		},
		{
			in:   "HTTPS://WWW.EXAMPLE.COM/",
			want: "https://www.example.com/",
		},
		{
			in:   "https://localhost/%61/foo",
			want: "https://localhost/a/foo",
		},
		{
			in:   "https://localhost/%aa/foo",
			want: "https://localhost/%AA/foo",
		},
		{
			in:   "https://localhost/%aa/%2F",
			want: "https://localhost/%AA/%2F",
		},
		{
			in:   "https://localhost/?q=%3D&a=b",
			want: "https://localhost/?q=%3D&a=b",
		},
		{
			in:   "https://localhost/?q=foo+bar+baz",
			want: "https://localhost/?q=foo%2Bbar%2Bbaz",
		},
		{
			in:   "https://localhost/a/./b/../c",
			want: "https://localhost/a/c",
		},
		{
			in:   "https://localhost/a//b//c/",
			want: "https://localhost/a/b/c/",
		},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			u, _ := url.Parse(tt.in)
			got, err := NormalizeURL(u)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.String() != tt.want {
				t.Errorf("NormalizeURL() = %v, want %v", got, tt.want)
			}
			t.Logf("%q -> %q", tt.in, got)
		})
	}
}

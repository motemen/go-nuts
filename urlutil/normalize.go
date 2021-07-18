package urlutil

import (
	"net/url"
	"strings"

	"golang.org/x/net/idna"
)

func CloneURL(u *url.URL) *url.URL {
	u2 := *u
	return &u2
}

var defaultPorts = map[string]string{
	"https": "443",
	"http":  "80",
}

// NormalizeURL normalizes URL u in such manner:
// - all components should be represented in ASCII
// - precent encoding in upper case
// https://datatracker.ietf.org/doc/html/rfc3986#section-6
func NormalizeURL(u *url.URL) (*url.URL, error) {
	u = CloneURL(u)

	u.Scheme = strings.ToLower(u.Scheme)

	port := u.Port()
	if port == defaultPorts[u.Scheme] {
		port = ""
	}

	hostname, err := idna.ToASCII(u.Hostname())
	if err != nil {
		return nil, err
	}
	hostname = strings.ToLower(hostname)

	u.Host = hostname
	if port != "" {
		u.Host += ":" + port
	}

	path := u.RawPath
	if path == "" {
		path = u.Path
	}
	if path == "" {
		path = "/"
	}

	path, err = normalizeComponent(path, "/", url.PathEscape)
	if err != nil {
		return nil, err
	}

	path = removeDotSegments(path)

	u.Path, err = url.PathUnescape(path)
	if err != nil {
		return nil, err
	}

	u.RawPath = ""

	u.RawQuery, err = normalizeComponent(u.RawQuery, "&;=", url.QueryEscape)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// https://datatracker.ietf.org/doc/html/rfc3986#section-5.2.4
func removeDotSegments(in string) string {
	segs := strings.Split(in, "/")
	result := make([]string, 0, len(segs))

	isAbsolute := false
	if segs[0] == "" {
		isAbsolute = true
		segs = segs[1:]
	}

	for _, seg := range segs {
		switch seg {
		case ".":
			// nop
		case "..":
			if len(result) > 0 {
				result = result[:len(result)-1]
			}
		default:
			result = append(result, seg)
		}
	}

	resultPath := strings.Join(result, "/")
	if isAbsolute {
		resultPath = "/" + resultPath
	}
	return resultPath
}

func normalizeComponent(component string, special string, escape func(string) string) (string, error) {
	escaped := ""

	for component != "" {
		var part, sep string
		if i := strings.IndexAny(component, special); i >= 0 {
			part, sep, component = component[:i], component[i:i+1], component[i+1:]
		} else {
			part, sep, component = component, "", ""
		}

		unescaped, err := url.PathUnescape(part)
		if err != nil {
			return "", err
		}

		escaped += escape(unescaped) + sep
	}

	return escaped, nil
}

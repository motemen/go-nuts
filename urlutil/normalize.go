package urlutil

import (
	"net/url"
	pathPkg "path"
	"strings"

	"golang.org/x/net/idna"
)

var defaultPorts = map[string]string{
	"https": "443",
	"http":  "80",
}

// NormalizeURL normalizes URL u in such manner:
// - all components should be represented in ASCII
// - precent encoding in upper case
func NormalizeURL(u *url.URL) (*url.URL, error) {
	u.Scheme = strings.ToLower(u.Scheme)

	port := u.Port()

	hostname, err := idna.ToASCII(u.Hostname())
	if err != nil {
		return nil, err
	}

	u.Host = strings.ToLower(hostname)

	if port == defaultPorts[u.Scheme] {
		port = ""
	}
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

	hadSlash := path[len(path)-1] == '/'
	u.RawPath = pathPkg.Clean(path)
	if hadSlash && u.RawPath[len(u.RawPath)-1] != '/' {
		u.RawPath += "/"
	}

	u.Path, err = url.PathUnescape(u.RawPath)
	if err != nil {
		return nil, err
	}

	u.RawQuery, err = normalizeComponent(u.RawQuery, "&;=", url.QueryEscape)
	if err != nil {
		return nil, err
	}

	return u, nil
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

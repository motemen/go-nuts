// Package giturl provides ParseGitURL which parses remote URLs under the way
// that git does.
package giturl

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	rxURLLike     = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9+.-]*://`)
	rxHostAndPort = regexp.MustCompile(`^([^:]+|\[.+?\]):([0-9]+)$`)
)

type Protocol int

const (
	ProtocolLocal Protocol = iota
	ProtocolFile
	ProtocolSSH
	ProtocolGit
)

func (p Protocol) String() string {
	switch p {
	case ProtocolLocal:
		return "file"
	case ProtocolFile:
		return "file"
	case ProtocolSSH:
		return "ssh"
	case ProtocolGit:
		return "git"
	default:
		panic("unreachable")
	}
}

var schemeToProtocol = map[string]Protocol{
	"ssh":     ProtocolSSH,
	"git":     ProtocolGit,
	"git+ssh": ProtocolSSH,
	"ssh+git": ProtocolSSH,
	"file":    ProtocolFile,
}

func ParseGitURL(giturl string) (proto Protocol, host string, port uint, path string, err error) {
	// ref: parse_connect_url() in connect.c

	if rxURLLike.MatchString(giturl) {
		var u *url.URL
		u, err = url.Parse(giturl)
		if err != nil {
			return
		}

		proto = schemeToProtocol[u.Scheme]
		host = u.Host
		if proto == ProtocolSSH {
			if m := rxHostAndPort.FindStringSubmatch(host); m != nil {
				var port64 uint64
				host = m[1]
				port64, err = strconv.ParseUint(m[2], 10, 16)
				if err != nil {
					return
				}
				port = uint(port64)
			}
		}
		if proto == ProtocolSSH && host[0] == '[' && host[len(host)-1] == ']' {
			host = host[1 : len(host)-1]
		}
		if u.User != nil {
			host = u.User.String() + "@" + host
		}
		path = u.Path
		if proto == ProtocolGit || proto == ProtocolSSH {
			if path[1] == '~' {
				path = path[1:]
			}
		} else if proto == ProtocolFile {
			host = ""
			path = u.Host + u.Path
		} else {
			panic("unreachable")
		}
	} else {
		colon := strings.IndexByte(giturl, ':')
		slash := strings.IndexByte(giturl, '/')

		if colon > -1 && (slash == -1 || colon < slash) /*&& !hasDosDrivePrefix(giturl)*/ {
			// For SCP-like URLs, colon must appear and be before any slashes
			// - user@host.xyz:path/to/repo.git/
			// - host.xyz:path/to/repo.git/
			// - user@[::1]:path/to/repo.git/
			// - [::1]:path/to/repo.git/
			proto = ProtocolSSH
			m := regexp.MustCompile(`^(.+?@)?\[(.+?)\]:(.*)`).FindStringSubmatch(giturl)
			if m != nil {
				host = m[1] + m[2]
				path = m[3]
			} else {
				host = giturl[:colon]
				path = giturl[colon+1:]
			}
			if path[1] == '~' {
				path = path[1:]
			}
		} else {
			path = giturl
		}
	}

	return
}

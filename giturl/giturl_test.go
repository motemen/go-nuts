package giturl

import "testing"

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

func checkParseGitURL(t *testing.T, url string) {
	t.Logf("URL: %s", url)

	proto, host, port, path, err := ParseGitURL(url)
	if err != nil {
		t.Fatal(err)
	}

	hostkey := "hostandport"
	if proto == "ssh" {
		hostkey = "userandhost"
	}

	portString := "NONE"
	if port != 0 {
		portString = fmt.Sprint(port)
	}

	got := fmt.Sprintf(`Diag: url=%s
Diag: protocol=%s
Diag: %s=%s
Diag: port=%s
Diag: path=%s
`, url, proto, hostkey, host, portString, path)

	if proto != "ssh" {
		got = regexp.MustCompile("(?m)^Diag: port=.*\n").ReplaceAllString(got, "")
	}

	expected, err := exec.Command("git", "fetch-pack", "--diag-url", url).Output()
	if err != nil {
		t.Fatal(err)
	}

	if got != string(expected) {
		t.Errorf(`URL %q failed:
# Got
%s
# Expected
%s
`, url, got, expected)
	}
}

func TestMain(m *testing.M) {
	cmd := exec.Command("git", "version")
	b, err := cmd.Output()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	fmt.Fprintf(os.Stderr, "# %s\n", b[:len(b)-1])
	os.Exit(m.Run())
}

// source: t/t5500-fetch-pack.sh
func TestParseGitURL(t *testing.T) {
	for _, repo := range []string{"repo", "re:po", "re/po"} {
		for _, proto := range []string{"ssh+git", "git+ssh", "git", "ssh"} {
			for _, host := range []string{"host", "user@host", "user@[::1]", "user@::1"} {
				checkParseGitURL(t, fmt.Sprintf("%s://%s/%s", proto, host, repo))
				checkParseGitURL(t, fmt.Sprintf("%s://%s/~%s", proto, host, repo))
			}
			for _, host := range []string{"host", "User@host", "User@[::1]"} {
				checkParseGitURL(t, fmt.Sprintf("%s://%s:22/%s", proto, host, repo))
			}
		}
		for _, proto := range []string{"file"} {
			checkParseGitURL(t, fmt.Sprintf("%s:///%s", proto, repo))
			checkParseGitURL(t, fmt.Sprintf("%s:///~%s", proto, repo))
		}
		for _, host := range []string{"nohost", "nohost:12", "[::1]", "[::1]:23", "[", "[:aa"} {
			checkParseGitURL(t, fmt.Sprintf("./%s:%s", host, repo))
			checkParseGitURL(t, fmt.Sprintf("./:%s/~%s", host, repo))
		}
		for _, host := range []string{"host", "[::1]"} {
			checkParseGitURL(t, fmt.Sprintf("%s:%s", host, repo))
			checkParseGitURL(t, fmt.Sprintf("%s:/~%s", host, repo))
		}
	}
}

func TestParseGitURL_ExtraSCPLike(t *testing.T) {
	for _, repo := range []string{"repo", "re:po", "re/po"} {
		for _, host := range []string{"user@host", "user@[::1]"} {
			checkParseGitURL(t, fmt.Sprintf("%s:%s", host, repo))
			checkParseGitURL(t, fmt.Sprintf("%s:/~%s", host, repo))
		}
	}
}

func TestParseGitURL_HTTP(t *testing.T) {
	for _, repo := range []string{"repo", "re:po", "re/po"} {
		for _, host := range []string{"host", "host:80"} {
			for _, proto := range []string{"http", "https"} {
				p, _, _, _, err := ParseGitURL(fmt.Sprintf("%s://%s/%s", proto, host, repo))
				if err != nil {
					t.Error(err)
				} else if p != proto {
					t.Errorf("expected protocol %q but got %q", proto, p)
				}
			}
		}
	}
}

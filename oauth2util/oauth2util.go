package oauth2util

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
)

// Config encapsulates typical OAuth2 authorization flow:
//   1. Try to restore previously-saved token,
//   2. If not available, start a local server for receiving code and prompt its URL,
//   3. Obtain an access token when code is received,
//   4. Store the token for later use.
type Config struct {
	// Required
	OAuth2Config *oauth2.Config

	// Required if TokenFile is empty
	Name string

	AuthCodeOptions []oauth2.AuthCodeOption

	// Defaults to <UserCacheDir>/<Name>/token.json
	TokenFile string
}

func (c *Config) DeleteTokenFile() (string, error) {
	tokenFile, err := c.getTokenFile()
	if err != nil {
		return "", err
	}

	return tokenFile, os.Remove(tokenFile)
}

func (c *Config) getTokenFile() (string, error) {
	if c.TokenFile != "" {
		return c.TokenFile, nil
	}

	cacheDirBase, err := os.UserCacheDir()
	if err != nil {
		return "", (fmt.Errorf("os.UserCacheDir: %w", err))
	}

	c.TokenFile = filepath.Join(cacheDirBase, c.Name, "token.json")

	return c.TokenFile, nil
}

// CreateOAuth2Client handles a typical authorization flow. See Config.
func (c *Config) CreateOAuth2Client(ctx context.Context) (*http.Client, error) {
	token, err := c.restoreToken()
	if err != nil {
		token, err = c.AuthorizeByTemporaryServer(
			ctx,
			func(authURL string) error {
				fmt.Printf("Visit below to authorize:\n%s\n", authURL)
				return nil
			},
		)
		if err != nil {
			return nil, err
		}

		err = c.storeToken(token)
		if err != nil {
			return nil, err
		}
	}

	return c.OAuth2Config.Client(ctx, token), nil
}

type CodeReceiver struct {
	ch    chan string
	State string
	*httptest.Server
}

func (c CodeReceiver) Code() <-chan string {
	return c.ch
}

func NewCodeReceiver() (*CodeReceiver, error) {
	ch := make(chan string)

	buf := make([]byte, 16)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}

	state := fmt.Sprintf("%x", sha256.Sum256(buf))

	s := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/favicon.ico" {
				http.Error(w, "Not Found", 404)
				return
			}

			if code := r.FormValue("code"); code != "" {
				if r.FormValue("state") != state {
					http.Error(w, "State mismatch", 400)
					return
				}

				w.Header().Set("Content-Type", "text/plain")
				fmt.Fprintln(w, "Authorized.")
				ch <- code
				return
			}
		}))

	return &CodeReceiver{
		ch:     ch,
		State:  state,
		Server: s,
	}, nil
}

func (c *Config) AuthorizeByTemporaryServer(ctx context.Context, prompt func(url string) error) (*oauth2.Token, error) {
	recv, err := NewCodeReceiver()
	if err != nil {
		return nil, err
	}

	defer recv.Close()

	var oauth2ConfigCopy oauth2.Config = *c.OAuth2Config
	oauth2ConfigCopy.RedirectURL = recv.URL

	err = prompt(oauth2ConfigCopy.AuthCodeURL(recv.State, c.AuthCodeOptions...))
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case code := <-recv.Code():
		return oauth2ConfigCopy.Exchange(ctx, code, c.AuthCodeOptions...)
	}
}

func (c *Config) restoreToken() (*oauth2.Token, error) {
	tokenFile, err := c.getTokenFile()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(tokenFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var token oauth2.Token
	err = json.NewDecoder(f).Decode(&token)
	return &token, err
}

func (c *Config) storeToken(token *oauth2.Token) error {
	tokenFile, err := c.getTokenFile()
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(tokenFile), 0o777)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(tokenFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(token)
}

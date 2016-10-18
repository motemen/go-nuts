package ctxlog

import (
	"log"
	"os"
)

var Logger = log.New(os.Stderr, "", log.LstdFlags)

type contextKey struct {
	name string
}

var PrefixContextKey = &contextKey{"prefix"}

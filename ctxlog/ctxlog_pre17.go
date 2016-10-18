// +build !appengine,!go1.7

package ctxlog

import (
	"golang.org/x/net/context"
)

func PrefixFromContext(ctx context.Context) string {
	prefix, _ := ctx.Value(PrefixContextKey).(string)
	return prefix
}

func NewContextWithPrefix(ctx context.Context, prefix string) context.Context {
	basePrefix := PrefixFromContext(ctx)
	return context.WithValue(ctx, PrefixContextKey, basePrefix+prefix)
}

func logf(ctx context.Context, level string, format string, args ...interface{}) {
	prefix := PrefixFromContext(ctx)
	args = append([]interface{}{prefix, level}, args...)
	Logger.Printf("%s%s: "+format, args...)
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, "debug", format, args...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, "info", format, args...)
}

func Warningf(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, "warning", format, args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, "error", format, args...)
}

func Criticalf(ctx context.Context, format string, args ...interface{}) {
	logf(ctx, "critical", format, args...)
}

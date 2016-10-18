// +build appengine

package ctxlog

import (
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

func logf(ctx context.Context, level string, format string, args ...interface{}) {
	prefix := PrefixFromContext(ctx)
	args = append([]interface{}{prefix, level}, args...)
	Logger.Printf("%s%s: "+format, args...)
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	prefix := PrefixFromContext(ctx)
	args = append([]interface{}{prefix}, args...)
	log.Debugf(ctx, "%s"+format, args...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	prefix := PrefixFromContext(ctx)
	args = append([]interface{}{prefix}, args...)
	log.Infof(ctx, "%s"+format, args...)
}

func Warningf(ctx context.Context, format string, args ...interface{}) {
	prefix := PrefixFromContext(ctx)
	args = append([]interface{}{prefix}, args...)
	log.Warningf(ctx, "%s"+format, args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	prefix := PrefixFromContext(ctx)
	args = append([]interface{}{prefix}, args...)
	log.Errorf(ctx, "%s"+format, args...)
}

func Criticalf(ctx context.Context, format string, args ...interface{}) {
	prefix := PrefixFromContext(ctx)
	args = append([]interface{}{prefix}, args...)
	log.Criticalf(ctx, "%s"+format, args...)
}

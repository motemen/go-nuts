// +build go1.7

package ctxlog

import (
	"testing"

	"context"
)

func testPrefix(t *testing.T, ctx context.Context, expect string) {
	logger := FromContext(ctx)
	if logger.Prefix() != expect {
		t.Errorf("prefix should be %q: got %q", expect, logger.Prefix())
	}
}

func TestFromContext(t *testing.T) {
	logger := FromContext(context.Background())
	if logger.Prefix() != "" {
		t.Errorf("prefix should be %q: got %q", "", logger.Prefix())
	}
}

func TestNewContext(t *testing.T) {
	ctx := NewContext(context.Background(), "prefix: ")
	testPrefix(t, ctx, "prefix: ")

	{
		ctx := NewContext(ctx, "foo: ")
		testPrefix(t, ctx, "prefix: foo: ")
	}

	testPrefix(t, ctx, "prefix: ")
}

package gocache

import (
	"testing"

	"github.com/bradberger/gocache/cache"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	c := New(0, 0)
	assert.NotNil(t, c)
	assert.NotNil(t, c.gc)
	assert.Equal(t, c.gc, c.Cache())
}

func TestGet(t *testing.T) {
	var v interface{}
	c := New(0, 0)
	assert.Equal(t, cache.ErrCacheMiss, c.Get("foobar", &v))

	var s string
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.NoError(t, c.Get("foo", &s))
	assert.Equal(t, "bar", s)

	// Invalid dst
	assert.Equal(t, cache.ErrInvalidDstVal, c.Get("foo", s))
}

func TestExists(t *testing.T) {
	c := New(0, 0)
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.True(t, c.Exists("foo"))
	assert.False(t, c.Exists("bar"))
}

func TestDel(t *testing.T) {
	c := New(0, 0)
	assert.NoError(t, c.Del("foobar"))
}

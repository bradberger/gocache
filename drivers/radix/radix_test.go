package radix

import (
	"testing"
	"time"

	"github.com/bradberger/gocache/cache"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	c := New()
	assert.NotNil(t, c)
	assert.NotNil(t, c.tree)
	assert.Equal(t, c.tree, c.Tree())
}

func TestGet(t *testing.T) {
	var v interface{}
	c := New()
	assert.Equal(t, cache.ErrCacheMiss, c.Get("foobar", &v))

	var s string
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.NoError(t, c.Get("foo", &s))
	assert.Equal(t, "bar", s)

	// Invalid dst
	assert.Equal(t, cache.ErrInvalidDstVal, c.Get("foo", s))
}

func TestExists(t *testing.T) {
	c := New()
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.True(t, c.Exists("foo"))
	assert.False(t, c.Exists("bar"))
}

func TestDel(t *testing.T) {
	c := New()
	assert.NoError(t, c.Del("foobar"))
}

func TestExpires(t *testing.T) {
	c := New()
	assert.NoError(t, c.Set("foo", "bar", 2*time.Second))
	time.Sleep(2001 * time.Millisecond)
	assert.False(t, c.Exists("foo"))
}

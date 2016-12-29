package twoqueue

import (
	"testing"
	"time"

	"github.com/bradberger/gocache/cache"
	"github.com/stretchr/testify/assert"
)

func TestInvalidNew(t *testing.T) {
	_, err := New(-1)
	assert.Error(t, err)
}

func TestSet(t *testing.T) {
	c, err := New(128)
	assert.NoError(t, err)
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.Error(t, c.Set("foo", "bar", time.Hour))
}

func TestGet(t *testing.T) {
	var v string
	c, _ := New(128)
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.True(t, c.Exists("foo"))
	assert.NoError(t, c.Get("foo", &v))
	assert.Equal(t, cache.ErrNotFound, c.Get("foobar", &v))
	assert.False(t, c.Exists("foobar"))
	assert.Equal(t, "bar", v)
	assert.NoError(t, c.Del("foo"))
	assert.False(t, c.Exists("foo"))
}

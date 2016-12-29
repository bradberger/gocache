package bigcache

import (
	"testing"
	"time"

	"github.com/allegro/bigcache"
	"github.com/bradberger/gocache/cache"
	"github.com/bradberger/gocache/codec"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Foo string
}

func TestNewWithClient(t *testing.T) {
	bc, _ := bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Minute))
	c := NewWithClient(bc)
	assert.NotNil(t, c)
	assert.NotNil(t, c.bc)
}

func TestGetSet(t *testing.T) {
	var v string
	c := New()
	assert.NotNil(t, c)
	assert.NotNil(t, c.bc)
	assert.False(t, c.Exists("foo"))
	assert.Equal(t, cache.ErrNotFound, c.Get("foo", &v))
	assert.Equal(t, cache.ErrUnsupportedAction, c.Set("foo", "bar", time.Second))
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.True(t, c.Exists("foo"))
	assert.NoError(t, c.Get("foo", &v))
	assert.Equal(t, cache.ErrUnsupportedAction, c.Del("foo"))
}

func TestSetErr(t *testing.T) {

	c := New()
	v := testStruct{"bar"}
	origCodec := Codec
	defer func() {
		Codec = origCodec
	}()
	Codec.Marshal = codec.ErrTestCodec.Marshal
	assert.EqualError(t, c.Set("foo", v, 0), "test error")
}

func TestGetErr(t *testing.T) {
	c := New()
	v := testStruct{"bar"}
	assert.NoError(t, c.Set("foo", &v, 0))
	origCodec := Codec
	defer func() {
		Codec = origCodec
	}()
	Codec.Unmarshal = codec.ErrTestCodec.Unmarshal
	assert.EqualError(t, c.Get("foo", &v), "test error")

}

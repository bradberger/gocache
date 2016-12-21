package redis

import (
	"testing"
	"time"

	"github.com/bradberger/gocache/cache"
	"github.com/bradberger/gocache/codec"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Foo string
}

func TestNew(t *testing.T) {
	r := New(":6379", 10)
	assert.NotNil(t, r.pool)
	assert.NotNil(t, r.Pool())
	assert.Equal(t, r.pool, r.Pool())
	assert.NoError(t, r.Close())
}

func TestSet(t *testing.T) {
	var v testStruct
	r := New(":6379", 10)
	assert.NoError(t, r.Set("foo", &testStruct{"bar"}, 0))
	assert.True(t, r.Exists("foo"))
	assert.NoError(t, r.Get("foo", &v))
	assert.Equal(t, "bar", v.Foo)
	assert.NoError(t, r.Del("foo"))
	assert.False(t, r.Exists("foo"))
	assert.Equal(t, cache.ErrCacheMiss, r.Get("foo", &v))
	assert.NoError(t, r.Close())
}

func TestSetExp(t *testing.T) {
	r := New(":6379", 10)
	assert.NoError(t, r.Set("fooexp", "bar", 1*time.Second))
	assert.True(t, r.Exists("fooexp"))
	time.Sleep(1001 * time.Millisecond)
	assert.False(t, r.Exists("fooexp"))
}

func TestSetErr(t *testing.T) {
	r := New(":6379", 10)
	origCodec := Codec
	defer func() {
		Codec = origCodec
	}()
	Codec = codec.ErrTestCodec
	assert.EqualError(t, r.Set("foo", &testStruct{"bar"}, 0), "test error")
}

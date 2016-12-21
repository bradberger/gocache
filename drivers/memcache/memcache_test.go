package memcache

import (
	"testing"

	"github.com/bradberger/gocache/codec"

	"github.com/bradberger/gocache/cache"
	"github.com/stretchr/testify/assert"
)

var (
	m = New(":11211")
)

type testStruct struct {
	Foo string
}

func TestNew(t *testing.T) {
	assert.NotNil(t, m)
	assert.NotNil(t, m.mc)
	assert.Equal(t, m.mc, m.Client())
}

func TestSet(t *testing.T) {
	v := testStruct{"bar"}
	assert.NoError(t, m.Set("foo", v, 0))
}

func TestAdd(t *testing.T) {
	var v string
	assert.NoError(t, m.Del("foo"))
	assert.NoError(t, m.Add("foo", "bar", 0))
	assert.NoError(t, m.Get("foo", &v))
	assert.Equal(t, "bar", v)
}

func TestAddErr(t *testing.T) {
	assert.NoError(t, m.Set("foo", "bar", 0))
	assert.Error(t, m.Add("foo", "bar", 0))
}

func TestAddCodecErr(t *testing.T) {
	origCodec := Codec
	defer func() {
		Codec = origCodec
	}()
	Codec.Marshal = codec.ErrTestCodec.Marshal
	assert.NoError(t, m.Del("foo"))
	assert.EqualError(t, m.Add("foo", "bar", 0), "test error")
}

func TestReplace(t *testing.T) {
	assert.NoError(t, m.Set("foo", "bar", 0))
	assert.NoError(t, m.Replace("foo", "bar", 0))
}

func TestReplaceErr(t *testing.T) {
	assert.NoError(t, m.Del("foo"))
	assert.Error(t, m.Replace("foo", "bar", 0))
}

func TestReplaceCodecErr(t *testing.T) {
	origCodec := Codec
	defer func() {
		Codec = origCodec
	}()
	assert.NoError(t, m.Set("foo", "bar", 0))
	Codec.Marshal = codec.ErrTestCodec.Marshal
	assert.EqualError(t, m.Replace("foo", "bar", 0), "test error")
}

func TestSetErr(t *testing.T) {

	v := testStruct{"bar"}
	origCodec := Codec
	defer func() {
		Codec = origCodec
	}()
	Codec.Marshal = codec.ErrTestCodec.Marshal
	assert.EqualError(t, m.Set("foo", v, 0), "test error")
}

func TestGet(t *testing.T) {
	v := testStruct{"bar"}
	vv := testStruct{}
	assert.NoError(t, m.Set("foo", v, 0))
	assert.NoError(t, m.Get("foo", &vv))
	assert.Equal(t, v, vv)
}

func TestGetDel(t *testing.T) {
	v := testStruct{"bar"}
	assert.NoError(t, m.Set("foo", v, 0))
	assert.True(t, m.Exists("foo"))
	assert.NoError(t, m.Del("foo"))
	assert.Equal(t, cache.ErrCacheMiss, m.Get("foo", &v))
	assert.False(t, m.Exists("foo"))
}

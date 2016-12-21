package memcache

import (
	"os"
	"testing"

	"github.com/bradberger/gocache/codec"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
	"google.golang.org/appengine/aetest"
)

var (
	ctx  context.Context
	done func()
	c    *Cache
)

type testStruct struct {
	Foo string
}

func TestMain(m *testing.M) {
	var err error
	ctx, done, err = aetest.NewContext()
	if err != nil {
		// log.Fatalf("Could not get test context: %v", err)
		ctx, done = context.WithCancel(context.Background())
	}
	code := m.Run()
	done()
	// If stderr/stdout are not closed tests will hang without -test.v
	os.Stdout.Sync()
	os.Stdout.Close()
	os.Stderr.Sync()
	os.Stderr.Close()
	os.Exit(code)
}

func TestNew(t *testing.T) {
	mc := New(ctx)
	assert.NotNil(t, mc)
	assert.NotNil(t, mc.ctx)
	assert.Equal(t, ctx, mc.ctx)
	assert.Equal(t, ctx, mc.Context())
}

func TestGetSet(t *testing.T) {
	v := testStruct{}
	mc := New(ctx)
	assert.NoError(t, mc.Del("foo"))
	assert.False(t, mc.Exists("foo"))
	assert.NoError(t, mc.Set("foo", testStruct{"bar"}, 0))
	assert.True(t, mc.Exists("foo"))
	assert.NoError(t, mc.Get("foo", &v))
	assert.NoError(t, mc.Del("foo"))
	assert.Equal(t, "bar", v.Foo)
}

func TestSetErr(t *testing.T) {
	mc := New(ctx)
	origCodec := Codec
	defer func() {
		Codec = origCodec
	}()
	Codec.Marshal = codec.ErrTestCodec.Marshal
	assert.Error(t, mc.Set("foo", &testStruct{"bar"}, 0))
}

func TestGetErr(t *testing.T) {
	var v testStruct
	mc := New(ctx)
	origCodec := Codec
	defer func() {
		Codec = origCodec
	}()
	Codec.Unmarshal = codec.ErrTestCodec.Unmarshal
	assert.NoError(t, mc.Set("foo", &testStruct{"bar"}, 0))
	assert.Error(t, mc.Get("foo", &v))
}

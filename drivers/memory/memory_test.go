package memory

import (
	"testing"
	"time"

	"github.com/bradberger/gocache/cache"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Foo string
}

func TestNew(t *testing.T) {
	c := New()
	assert.NotNil(t, c)
	assert.NotNil(t, c.m)
	assert.NotNil(t, c.exp)
}

func TestGetCold(t *testing.T) {
	c := New()
	var v testStruct
	assert.Equal(t, cache.ErrCacheMiss, c.Get("foobar", &v))
}

func TestSet(t *testing.T) {
	s := "foo"
	c := New()
	v := testStruct{"bar"}
	assert.NoError(t, c.Set("foo", &v, 0))
	assert.True(t, c.Exists("foo"))
	assert.Error(t, c.Set("foo", &s, 0))
	assert.NoError(t, c.Set("foo", testStruct{"foo"}, 0))
	assert.NoError(t, c.Get("foo", &v))
	assert.Equal(t, "foo", v.Foo)

}

func TestGetNoPtr(t *testing.T) {
	c := New()
	v := testStruct{"bar"}
	var vv testStruct
	assert.NoError(t, c.Set("foo", v, 0))
	assert.Equal(t, cache.ErrInvalidDstVal, c.Get("foo", vv))
}

func TestGet(t *testing.T) {
	c := New()
	v := testStruct{"bar"}
	var vv, vvv testStruct
	var s string
	assert.NoError(t, c.Set("foo", v, 0))
	assert.NoError(t, c.Get("foo", &vv))

	assert.NoError(t, c.Set("foo", &vv, 0))
	assert.NoError(t, c.Get("foo", &vvv))
	assert.Equal(t, ErrCannotAssignValue, c.Get("foo", &s))
	assert.EqualValues(t, v, vv)
	assert.EqualValues(t, vv, vvv)
}

func TestGetSlice(t *testing.T) {
	c := New()
	v := []testStruct{{"bar"}}
	var vv []testStruct
	assert.NoError(t, c.Set("foo", v, 0))
	assert.NoError(t, c.Get("foo", &vv))
	assert.EqualValues(t, v, vv)
}

func TestGetInt64(t *testing.T) {
	c := New()
	v := int64(1)
	var vv int64
	assert.NoError(t, c.Set("foo", v, 0))
	assert.NoError(t, c.Get("foo", &vv))
	assert.Equal(t, int64(1), vv)
	assert.Equal(t, v, vv)
}

func TestExpires(t *testing.T) {
	c := New()
	v := 1
	assert.NoError(t, c.Set("foo", v, 1*time.Second))
	_, exists := c.exp["foo"]
	assert.True(t, exists)
	assert.True(t, c.Exists("foo"))
	time.Sleep(2 * time.Second)
	assert.False(t, c.Exists("foo"))
}

func BenchmarkSet(b *testing.B) {
	c := New()
	for i := 0; i < b.N; i++ {
		c.Set("foo", "bar", 0)
	}
}

func BenchmarkSetConcurrent(b *testing.B) {
	c := New()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Set("foo", "bar", 0)
		}
	})
}

func BenchmarkGet(b *testing.B) {
	c := New()
	c.Set("foo", "bar", 0)
	for i := 0; i < b.N; i++ {
		var s string
		c.Get("foo", &s)
	}
}

func BenchmarkGetConcurrent(b *testing.B) {
	c := New()
	c.Set("foo", "bar", 0)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var s string
			c.Get("foo", &s)
		}
	})
}

func BenchmarkGetSetConcurrent(b *testing.B) {
	c := New()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var s string
			c.Set("foo", "bar", 0)
			c.Get("foo", &s)
		}
	})
}

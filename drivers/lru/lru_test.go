package lru

import (
	"testing"
	"time"

	"github.com/bradberger/gocache/cache"
	"github.com/stretchr/testify/assert"
)

type testStruct struct {
	Foo string
}

type diffTestStruct struct {
	Bar string
}

func TestNewBasic(t *testing.T) {
	c := NewBasic(Megabyte, 100)
	assert.NotNil(t, c)
	assert.NotNil(t, c.cache)
	assert.NotNil(t, c.ll)
}

func TestSet(t *testing.T) {
	c := NewBasic(Megabyte, 100)
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.NoError(t, c.Set("foobar", "", 0))
	assert.NoError(t, c.Set("bar", "foo", 0))
	assert.Equal(t, "bar", c.ll.Front().Value.(*item).key)
	assert.Equal(t, "foo", c.ll.Back().Value.(*item).key)
	assert.True(t, c.Exists("foo"))
	assert.True(t, c.Exists("bar"))
	assert.True(t, c.Exists("foobar"))
	assert.Equal(t, int(3), c.Len())

	// Make sure non-front elements at front after set again
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.Equal(t, "foo", c.ll.Front().Value.(*item).key)
}

func TestSetWithStructs(t *testing.T) {
	c := NewBasic(Megabyte, 100)
	var v testStruct

	// Assign values back to back with different types, but not pointers
	assert.NoError(t, c.Set("foo", testStruct{"bar"}, 0))
	assert.NoError(t, c.Set("foo", diffTestStruct{"bar"}, 0))

	// Try to assign a different type to existing pointer.
	assert.NoError(t, c.Set("foo", &testStruct{"bar"}, 0))
	assert.NoError(t, c.Set("foo", diffTestStruct{"bar"}, 0))

	// Try to assign same type, non-pointer to existing pointer.
	assert.NoError(t, c.Set("foo", testStruct{"foobar"}, 0))
	assert.NoError(t, c.Get("foo", &v))
	assert.Equal(t, "foobar", v.Foo)

	// Try to get with a different dstVal
	assert.Equal(t, ErrCannotAssignValue, c.Get("foo", &diffTestStruct{}))
}

func TestSetItem(t *testing.T) {
	c := NewBasic(Megabyte, 100)

	assert.Equal(t, cache.ErrNotStored, c.CompareAndSwap("foo", "bar", 0, 1))
	assert.NoError(t, c.SetItem("foo", "bar", 0, 1))
	assert.NoError(t, c.CompareAndSwap("foo", "bar", 0, 1))
	assert.NoError(t, c.SetItem("foo", "bar", 0, 2))
	assert.Equal(t, cache.ErrCASConflict, c.CompareAndSwap("foo", "bar", 0, 1))
}

func TestExpire(t *testing.T) {
	c := NewBasic(Megabyte, 100)
	assert.NotPanics(t, func() {
		c.expire("foobar")
	})
}

func TestGet(t *testing.T) {
	var s string
	c := NewBasic(Megabyte, 100)
	assert.Equal(t, cache.ErrCacheMiss, c.Get("foo", &s))

	ss := "bar"
	assert.NoError(t, c.Set("foo", ss, 0))
	assert.NoError(t, c.Get("foo", &s))
	assert.Equal(t, s, ss)
}

func TestTouch(t *testing.T) {
	c := NewBasic(Megabyte, 100)
	assert.Equal(t, cache.ErrCacheMiss, c.Touch("foo", 0))
	assert.NoError(t, c.Set("foo", "bar", time.Minute))
	assert.NoError(t, c.Set("bar", "foo", 0))
	assert.NoError(t, c.Touch("foo", time.Hour))
	assert.Equal(t, "foo", c.ll.Front().Value.(*item).key)
	assert.True(t, time.Now().Add(30*time.Minute).Before(c.ll.Front().Value.(*item).exp))
	assert.NoError(t, c.Touch("foo", 0))
	assert.True(t, c.ll.Front().Value.(*item).exp.IsZero())
}

func TestGetExpired(t *testing.T) {
	var v string
	c := NewBasic(Megabyte, 100)
	assert.NoError(t, c.Set("foo", "bar", -1*time.Minute))
	assert.Equal(t, cache.ErrCacheMiss, c.Get("foo", &v))
}

func TestInvalidGet(t *testing.T) {
	var s string
	c := NewBasic(Megabyte, 100)
	assert.NoError(t, c.Set("foo", s, 0))
	assert.Equal(t, cache.ErrInvalidDstVal, c.Get("foo", s))
}

func TestSetRemove(t *testing.T) {
	c := NewBasic(Megabyte, 2)
	assert.NoError(t, c.Set("foobar", "", 0))
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.NoError(t, c.Set("bar", "foo", 0))
	assert.Equal(t, "bar", c.ll.Front().Value.(*item).key)
	assert.Equal(t, "foo", c.ll.Back().Value.(*item).key)
	assert.Equal(t, int(2), c.Len())
}

func TestDel(t *testing.T) {
	c := NewBasic(Megabyte, 2)
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.NoError(t, c.Del("foo"))
	assert.Nil(t, c.ll.Back())
	assert.Nil(t, c.ll.Front())
}

func TestExpires(t *testing.T) {
	c := NewBasic(Megabyte, 2)
	assert.NoError(t, c.Set("foo", "bar", 1*time.Second))
	time.Sleep(2 * time.Second)
	assert.False(t, c.Exists("foo"))
}

func TestTicker(t *testing.T) {
	c := NewBasic(Megabyte, 10)
	assert.NoError(t, c.Set("foo", "bar", 0))
}

func TestSizeEviction(t *testing.T) {
	c := NewBasic(Byte, 100)
	assert.Equal(t, uint64(0), c.size)
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.Equal(t, uint64(0), c.size)
	assert.False(t, c.Exists("foo"))
}

func TestAdd(t *testing.T) {
	c := NewBasic(Megabyte, 1000)
	assert.NoError(t, c.Add("foo", "bar", 0))
	assert.Equal(t, cache.ErrNotStored, c.Add("foo", "bar", 0))
}

func TestReplace(t *testing.T) {
	c := NewBasic(Megabyte, 1000)
	var v string
	assert.Equal(t, cache.ErrNotStored, c.Replace("foo", "bar"))
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.NoError(t, c.Set("bar", "foo", 0))
	assert.NoError(t, c.Replace("foo", "bar2"))
	assert.NoError(t, c.Get("foo", &v))
	assert.Equal(t, "bar2", v)
	assert.Equal(t, c.cache["foo"], c.ll.Front())
}

func TestIncrement(t *testing.T) {
	var i uint64
	c := NewBasic(Megabyte, 1000)

	assert.NoError(t, c.Set("foo", uint64(1), 0))
	v, err := c.Increment("foo", 1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), v)
	assert.NoError(t, c.Get("foo", &i))
	assert.Equal(t, uint64(2), i)

	// Increment non-existant value.
	v, err = c.Increment("bar", 1)
	assert.Equal(t, uint64(0), v)
	assert.Equal(t, cache.ErrNotStored, err)

	// Increment a string
	assert.NoError(t, c.Set("foobar", "foobar", 0))
	_, err = c.Increment("foobar", 1)
	assert.Equal(t, ErrCannotAssignValue, err)
}

func TestDecrement(t *testing.T) {
	var i uint64
	c := NewBasic(Megabyte, 1000)

	assert.NoError(t, c.Set("foo", uint64(2), 0))
	v, err := c.Decrement("foo", 1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), v)
	assert.NoError(t, c.Get("foo", &i))
	assert.Equal(t, uint64(1), i)

	// Increment non-existant value.
	v, err = c.Decrement("bar", 1)
	assert.Equal(t, uint64(0), v)
	assert.Equal(t, cache.ErrNotStored, err)

	// Increment a string
	assert.NoError(t, c.Set("foobar", "foobar", 0))
	_, err = c.Decrement("foobar", 1)
	assert.Equal(t, ErrCannotAssignValue, err)
}

func TestPrepend(t *testing.T) {

	c := NewBasic(Megabyte, 1000)

	var s string
	sbar := "bar"
	assert.Equal(t, ErrCannotAssignValue, c.Prepend("foo", 1))
	assert.Equal(t, cache.ErrNotStored, c.Prepend("foo", sbar))
	assert.NoError(t, c.Set("foo", sbar, 0))
	assert.NoError(t, c.Prepend("foo", &sbar))
	assert.NoError(t, c.Prepend("foo", "foo"))
	assert.NoError(t, c.Get("foo", &s))
	assert.Equal(t, "foobarbar", s)

	// Prepend with byte slices
	var b []byte
	bar := []byte("bar")
	assert.Equal(t, ErrCannotAssignValue, c.Prepend("bar", 1))
	assert.Equal(t, cache.ErrNotStored, c.Prepend("bar", []byte("bar")))
	assert.NoError(t, c.Set("bar", bar, 0))
	assert.NoError(t, c.Prepend("bar", &bar))
	assert.NoError(t, c.Prepend("bar", []byte("foo")))
	assert.NoError(t, c.Get("bar", &b))
	assert.Equal(t, []byte("foobarbar"), b)

	// Prepend with invalid types
	assert.NoError(t, c.Set("int", 1, 0))
	assert.Equal(t, ErrCannotAssignValue, c.Prepend("int", "foobar"))
}

func TestAppend(t *testing.T) {

	c := NewBasic(Megabyte, 1000)

	var s string
	sbar := "bar"
	assert.Equal(t, ErrCannotAssignValue, c.Append("foo", 1))
	assert.Equal(t, cache.ErrNotStored, c.Append("foo", sbar))
	assert.NoError(t, c.Set("foo", sbar, 0))
	assert.NoError(t, c.Append("foo", &sbar))
	assert.NoError(t, c.Append("foo", "foo"))
	assert.NoError(t, c.Get("foo", &s))
	assert.Equal(t, "barbarfoo", s)

	// Append with byte slices
	var b []byte
	bar := []byte("bar")
	assert.Equal(t, ErrCannotAssignValue, c.Append("bar", 1))
	assert.Equal(t, cache.ErrNotStored, c.Append("bar", bar))
	assert.NoError(t, c.Set("bar", bar, 0))
	assert.NoError(t, c.Append("bar", "foo"))
	assert.NoError(t, c.Append("bar", &bar))
	assert.NoError(t, c.Get("bar", &b))
	assert.Equal(t, []byte("barfoobar"), b)

	// Append with invalid types
	assert.NoError(t, c.Set("int", 1, 0))
	assert.Equal(t, ErrCannotAssignValue, c.Append("int", "foobar"))
}

func BenchmarkSet(b *testing.B) {
	c := NewBasic(Gigabyte, b.N)
	for i := 0; i < b.N; i++ {
		c.Set(randKey(), "bar", 0)
	}
}

func BenchmarkSetParallel(b *testing.B) {
	c := NewBasic(Gigabyte, b.N)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Set(randKey(), "bar", 0)
		}
	})
}

func BenchmarkGet(b *testing.B) {
	c := NewBasic(Gigabyte, 500000)
	for i := 0; i < b.N; i++ {
		var s string
		c.Get(randKey(), &s)
	}
}

func BenchmarkGetParallel(b *testing.B) {
	c := NewBasic(Gigabyte, 500000)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var s string
			c.Get(randKey(), &s)
		}
	})
}

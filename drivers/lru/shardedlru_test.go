package lru

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/bradberger/gocache/cache"
	"github.com/stretchr/testify/assert"
)

func randKey() string {
	return strconv.Itoa(rand.Intn(10000000))
}

func TestShardedGetSet(t *testing.T) {
	c := NewSharded(2<<10, Megabyte, 256)

	var v string
	assert.False(t, c.Exists("foo"))
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.NoError(t, c.Touch("foo", 0))
	assert.Error(t, c.Add("foo", "bar", 0))
	assert.NoError(t, c.Replace("foo", "bar"))
	assert.NoError(t, c.Prepend("foo", "foo"))
	assert.NoError(t, c.Append("foo", "foo"))
	assert.True(t, c.Exists("foo"))
	assert.NoError(t, c.Get("foo", &v))
	assert.NoError(t, c.Del("foo"))
	assert.False(t, c.Exists("foo"))
	assert.Equal(t, "foobarfoo", v)

	assert.NoError(t, c.Set("ct", uint64(2), 0))
	ct, err := c.Increment("ct", 10)

	assert.Equal(t, uint64(12), ct)
	assert.NoError(t, err)

	ct, err = c.Decrement("ct", 1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(11), ct)

	assert.NoError(t, c.Get("ct", &ct))
	assert.Equal(t, uint64(11), ct)
}

func TestShardedSetItem(t *testing.T) {
	c := NewSharded(2<<10, Megabyte, 256)

	assert.Equal(t, cache.ErrNotStored, c.CompareAndSwap("foo", "bar", 0, 1))
	assert.NoError(t, c.SetItem("foo", "bar", 0, 1))

	var v string
	cas, err := c.GetItem("foo", &v)
	assert.Equal(t, uint64(1), cas)
	assert.NoError(t, err)

	assert.NoError(t, c.CompareAndSwap("foo", "bar", 0, 1))
	assert.NoError(t, c.SetItem("foo", "bar", 0, 2))
	assert.Equal(t, cache.ErrCASConflict, c.CompareAndSwap("foo", "bar", 0, 1))
}

func BenchmarkGetSharded(b *testing.B) {
	b.StopTimer()
	c := NewSharded(2<<10, Megabyte, 256)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		var s string
		c.Get(randKey(), &s)
	}
}

func BenchmarkGetShardedParallel(b *testing.B) {
	b.StopTimer()
	c := NewSharded(2<<10, Megabyte, 256)
	b.StartTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var s string
			c.Get(randKey(), &s)
		}
	})
}

func BenchmarkSetSharded(b *testing.B) {
	b.StopTimer()
	c := NewSharded(2<<10, Megabyte, 256)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		c.Set(randKey(), "bar", 0)
	}
}

func BenchmarkSetShardedParallel(b *testing.B) {
	b.StopTimer()
	c := NewSharded(2<<10, Megabyte, 256)
	b.StartTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Set(randKey(), "bar", 0)
		}
	})
}

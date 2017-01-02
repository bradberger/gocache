package btree

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/bradberger/gocache/cache"
	"github.com/stretchr/testify/assert"
)

func TestShards(t *testing.T) {
	c := NewContainer(1024)
	var v string
	assert.Equal(t, 1024, len(c.ch.Members()))
	assert.False(t, c.Exists("foo"))
	assert.Equal(t, cache.ErrNotFound, c.Get("foo", &v))
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.True(t, c.Exists("foo"))
	assert.NoError(t, c.Get("foo", &v))
	assert.Equal(t, "bar", v)
	assert.NoError(t, c.Del("foo"))
	assert.False(t, c.Exists("foo"))
}

func TestGetSetDel(t *testing.T) {
	c := New()
	var v string
	assert.False(t, c.Exists("foo"))
	assert.Equal(t, cache.ErrNotFound, c.Get("foo", &v))
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.True(t, c.Exists("foo"))
	assert.NoError(t, c.Get("foo", &v))
	assert.Equal(t, "bar", v)
	assert.NoError(t, c.Del("foo"))
	assert.False(t, c.Exists("foo"))
}

func TestExpires(t *testing.T) {
	c := New()
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.NoError(t, c.Set("bar", "foo", 1*time.Millisecond))
	time.Sleep(100 * time.Millisecond)
	assert.False(t, c.Exists("bar"))
}

func TestNew(t *testing.T) {
	c := New()
	assert.NotNil(t, c.exp)
	assert.NotNil(t, c.tree)
	assert.Equal(t, c.tree, c.Tree())
}

func TestDereference(t *testing.T) {
	c := New()
	s := "bar"
	assert.NoError(t, c.Set("foo", &s, 0))
	assert.NoError(t, c.Set("foo", "foobar", 0))
	assert.Equal(t, "bar", s)
}

func BenchmarkGetParallel(b *testing.B) {
	c := New()
	c.Set("foo", "bar", 0)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var s string
			c.Get("foo", &s)
		}
	})
}

func BenchmarkSetParallel(b *testing.B) {
	c := New()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Set("foo", "bar", 0)
		}
	})
}

func BenchmarkGet1M(b *testing.B) {
	b.StopTimer()
	c := New()
	for i := 0; i < 1000000; i++ {
		c.Set(rnd.Get(), "bar", 0)
	}
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var v string
			c.Get(rnd.Get(), &v)
		}
	})
}

func BenchmarkGetSharded1M(b *testing.B) {
	b.StopTimer()
	c := NewContainer(1024)
	for i := 0; i < 1000000; i++ {
		c.Set(rnd.Get(), "bar", 0)
	}
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var v string
			c.Get(rnd.Get(), &v)
		}
	})
}

func BenchmarkSet1M(b *testing.B) {
	c := New()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Set(rnd.Get(), "bar", 0)
		}
	})
}

func BenchmarkSetSharded1M(b *testing.B) {
	b.StopTimer()
	c := NewContainer(1024)
	b.StartTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Set(rnd.Get(), "bar", 0)
		}
	})
}

var rnd *randomKey

func init() {
	rnd = &randomKey{pool: make([]string, 0)}
	for i := 0; i < 1000001; i++ {
		rnd.pool = append(rnd.pool, fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d", rand.Int())))))
	}
}

type randomKey struct {
	pool []string
	len  int
}

func (rk *randomKey) Get() string {
	return rk.pool[rand.Intn(1000000)]
}

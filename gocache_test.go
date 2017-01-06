package gocache

import (
	"testing"
	"time"

	"github.com/bradberger/gocache/cache"
	"github.com/bradberger/gocache/drivers/lru"
	"github.com/stretchr/testify/assert"
)

func TestReplicationN(t *testing.T) {

	var s string
	c := New()

	assert.Error(t, c.Set("foo", "bar", 0))

	c.AddNode("node-01", lru.NewBasic(lru.Megabyte, 2<<10))
	c.ReplicateToN(2)
	assert.NoError(t, c.Set("foo", "bar", 0))

	c.AddNode("node-02", lru.NewBasic(lru.Megabyte, 2<<10))
	assert.NoError(t, c.Set("foo", "bar", 0))

	assert.NoError(t, c.nodes["node-01"].Get("foo", &s))
	assert.NoError(t, c.nodes["node-02"].Get("foo", &s))
}

func TestSetNodes(t *testing.T) {
	c := New()
	assert.Error(t, c.ReplaceNode("node-01", lru.NewBasic(lru.Megabyte, 2<<10)))
	assert.Error(t, c.SetNode("node-01", nil))
	assert.NoError(t, c.AddNode("node-01", lru.NewBasic(lru.Megabyte, 2<<10)))
	assert.NoError(t, c.ReplaceNode("node-01", lru.NewBasic(lru.Megabyte, 2<<10)))
	assert.Error(t, c.AddNode("node-01", lru.NewBasic(lru.Megabyte, 2<<10)))
	assert.NoError(t, c.RemoveNode("node-01"))
	assert.Len(t, c.nodes, 0)
}

func TestGetErr(t *testing.T) {
	var v string
	c := New()
	c.AddNode("node-01", lru.NewBasic(lru.Megabyte, 2<<10))
	assert.Equal(t, cache.ErrCacheMiss, c.Get("foo", &v))
}

func TestReplicationSync(t *testing.T) {

	var s string
	c := New()
	c.AddNode("node-01", lru.NewBasic(lru.Megabyte, 2<<10))
	c.AddNode("node-02", lru.NewBasic(lru.Megabyte, 2<<10))
	c.ReplicateToN(2)
	c.SetReplicateMethod(ReplicateSync)

	assert.False(t, c.Exists("foo"))
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.True(t, c.Exists("foo"))
	assert.NoError(t, c.nodes["node-01"].Get("foo", &s))
	assert.NoError(t, c.nodes["node-02"].Get("foo", &s))
	assert.NoError(t, c.nodes["node-01"].Del("foo"))
	assert.True(t, c.Exists("foo"))
	assert.NoError(t, c.Get("foo", &s))
	assert.Equal(t, "bar", s)
	assert.NoError(t, c.nodes["node-02"].Del("foo"))
	assert.False(t, c.Exists("foo"))
	assert.NoError(t, c.Del("foo"))
}

func TestGetNErr(t *testing.T) {
	c := New()
	assert.Error(t, c.Del("foo"))
	assert.Error(t, c.Set("foo", "bar", 0))
	assert.Error(t, c.Get("foo", nil))
	assert.False(t, c.Exists("foo"))
}

func TestAsyncExists(t *testing.T) {
	c := New()
	c.AddNode("node-01", lru.NewBasic(lru.Megabyte, 2<<10))
	c.AddNode("node-02", lru.NewBasic(lru.Megabyte, 2<<10))
	c.ReplicateToN(2)
	c.SetReplicateMethod(ReplicateSync)
	c.SetRetrievalMethod(RetrieveAsync)

	assert.False(t, c.Exists("foo"))
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.True(t, c.Exists("foo"))
}

func TestSetTimeout(t *testing.T) {
	c := New()
	c.SetTimeout(1)
	assert.Equal(t, time.Nanosecond, c.timeoutSet)
}

func TestAsyncGet(t *testing.T) {
	c := New()
	c.AddNode("node-01", lru.NewBasic(lru.Megabyte, 2<<10))
	c.AddNode("node-02", lru.NewBasic(lru.Megabyte, 2<<10))
	c.ReplicateToN(2)
	c.SetReplicateMethod(ReplicateSync)
	c.SetRetrievalMethod(RetrieveAsync)

	var s string
	c.GetTimeout(time.Nanosecond)
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.Equal(t, ErrGetTimeout, c.Get("foo", &s))
	assert.NoError(t, c.Del("foo"))

	c.GetTimeout(0)
	assert.Equal(t, cache.ErrCacheMiss, c.Get("foo", &s))
	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.NoError(t, c.Get("foo", &s))
	assert.Equal(t, "bar", s)
}

func TestDelTimeout(t *testing.T) {
	c := New()
	c.AddNode("node-01", lru.NewBasic(lru.Megabyte, 2<<10))
	c.SetTimeout(time.Nanosecond)

	assert.NoError(t, c.Set("foo", "bar", 0))
	assert.Equal(t, ErrSetTimeout, c.Del("foo"))
}

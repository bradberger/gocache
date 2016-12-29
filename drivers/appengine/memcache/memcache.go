// Package memecache implements a cache driver for Google App Engine Go mememcache
package memcache

import (
	"time"

	"github.com/bradberger/gocache/cache"
	"github.com/bradberger/gocache/codec"
	"golang.org/x/net/context"
	ae "google.golang.org/appengine/memcache"
)

// Codec is the App Engine Memcache codec to use. By default it's Gob, but if you
// want to use JSON, just set it first before accessing the cache.
var (
	Codec = codec.Gob

	// ensure Memcache struct implements the cache.Cache interface
	_ cache.Cache = (*Cache)(nil)
)

// Cache is
type Cache struct {
	ctx context.Context
}

// New returns a new empty client initialized with the given context
func New(ctx context.Context) *Cache {
	return &Cache{ctx: ctx}
}

// Context returns the context associated with the client. If the context is nil,
// it will return a background context which will trigger an App Engine error but
// will prevent a panic.
func (c *Cache) Context() context.Context {
	if c.ctx == nil {
		return context.Background()
	}
	return c.ctx
}

// Set implements the "cache.Cache".Set() interface
func (c *Cache) Set(key string, value interface{}, expiration time.Duration) (err error) {
	b, err := Codec.Marshal(value)
	if err != nil {
		return
	}
	return ae.Set(c.Context(), &ae.Item{Key: key, Expiration: expiration, Value: b})
}

// Get implements the "cache.Cache".Get() interface
func (c *Cache) Get(key string, dstVal interface{}) (err error) {
	item, err := ae.Get(c.Context(), key)
	if err != nil {
		if err == ae.ErrCacheMiss {
			err = cache.ErrCacheMiss
		}
		return
	}
	return Codec.Unmarshal(item.Value, dstVal)
}

// Del implements the "cache.Cache".Del() interface
func (c *Cache) Del(key string) (err error) {
	err = ae.Delete(c.Context(), key)
	if err == ae.ErrCacheMiss {
		err = nil
	}
	return
}

// Exists implements the "cache.Cache".Exists() interface
func (c *Cache) Exists(key string) bool {
	return c.Get(key, nil) == nil
}

// WithContext sets the internal context for future actions using the given context
func (c *Cache) WithContext(ctx context.Context) *Cache {
	c.ctx = ctx
	return c
}

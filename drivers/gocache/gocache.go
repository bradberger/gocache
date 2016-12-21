// Package gocache is a wrapper for Patrick Mylund Nielsen's go-cache package fouund at
// github.com/patricknm/go-cache
package gocache

import (
	"reflect"
	"time"

	"github.com/bradberger/gocache/cache"
	gc "github.com/patrickmn/go-cache"
)

var (
	// ensure LRU struct implements the cache.KeyStore interface
	_ cache.Cache = (*Cache)(nil)
)

// Cache is the struct which implememnts the cache.Cache interface.
type Cache struct {
	gc *gc.Cache
}

// New returns a new go-cache struct
func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	return &Cache{gc.New(defaultExpiration, cleanupInterval)}
}

// Get implements the cache.Cache getter interface
func (c *Cache) Get(key string, dstVal interface{}) error {
	val, found := c.gc.Get(key)
	if !found {
		return cache.ErrCacheMiss
	}
	el := reflect.ValueOf(dstVal)
	if el.Kind() == reflect.Ptr {
		el = el.Elem()
	}
	if !el.CanSet() {
		return cache.ErrInvalidDstVal
	}
	el.Set(reflect.ValueOf(val))
	return nil
}

// Set implements the cache.Cache setter interface
func (c *Cache) Set(key string, value interface{}, exp time.Duration) error {
	// By default, go-cache uses -1 for NoExpiration, but we expect 0 to be no expiration.
	if exp == 0 {
		exp = -1
	}
	c.gc.Set(key, value, exp)
	return nil
}

// Exists implements the cache.Cache exists interface
func (c *Cache) Exists(key string) bool {
	_, exists := c.gc.Get(key)
	return exists
}

// Del implements the cache.Cache deletion interface
func (c *Cache) Del(key string) error {
	c.gc.Delete(key)
	return nil
}

// Cache returns the underlying go-cache struct so you can use it's more powerful feature-set.
func (c *Cache) Cache() *gc.Cache {
	return c.gc
}

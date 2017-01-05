// Package kv is a wrapper for permanent key/value drivers at github.com/bradberger/gokv
// This allows using permanent key/value datastores like BoltDB, DiskV, LevelDB, etc., for
// persistent caching. Even though the underlying datastore is permanent, this driver
// implements the expiration of keys.
package kv

import (
	"sync"
	"time"

	"github.com/bradberger/gocache/cache"
	"github.com/bradberger/gokv/kv"
)

var (
	_ cache.Cache = (*Client)(nil)
)

// Client is a client which implmenets the "cache.Cache" interface while connecting
// to a datastore which implements the "kv.Store" interface
type Client struct {
	store kv.Store
	exp   map[string]time.Time

	sync.RWMutex
}

// New creates a new client with the given permanent key/value store
func New(store kv.Store) *Client {
	return &Client{store: store, exp: make(map[string]time.Time, 0)}
}

// Get implements the "cache.Cache".Get() interface
func (c *Client) Get(key string, dstVal interface{}) (err error) {
	if err = c.store.Get(key, dstVal); err == kv.ErrNotFound {
		err = cache.ErrCacheMiss
	}
	return
}

// Set implements the "cache.Cache".Set() interface
func (c *Client) Set(key string, val interface{}, exp time.Duration) (err error) {
	if err = c.store.Set(key, val); err != nil {
		return err
	}
	if exp > 0 {
		c.Lock()
		c.exp[key] = time.Now().Add(exp)
		c.Unlock()
		go func(key string, exp time.Duration) {
			time.Sleep(exp)
			c.RLock()
			t := c.exp[key]
			c.RUnlock()
			if !t.IsZero() && t.Before(time.Now()) {
				c.Del(key)
			}
		}(key, exp)
	}
	return nil
}

// Del implements the "cache.Cache".Del() interface
func (c *Client) Del(key string) (err error) {
	if err = c.store.Del(key); err == nil {
		c.Lock()
		delete(c.exp, key)
		c.Unlock()
	}
	return
}

// Exists implements the "cache.Cache".Exists() interface
func (c *Client) Exists(key string) bool {
	var val interface{}
	return c.Get(key, val) == nil
}

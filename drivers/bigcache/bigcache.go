package bigcache

import (
	"sync"
	"time"

	"github.com/allegro/bigcache"
	"github.com/bradberger/gocache/cache"
	"github.com/bradberger/gocache/codec"
)

var (
	// Codec is the codec used to marshal/unmarshal interfaces into the byte slices required by the Bigcache client
	Codec = codec.Gob

	_ cache.Cache = (*Client)(nil)
)

type Client struct {
	bc *bigcache.BigCache
	sync.RWMutex
}

// New returns a new client with the default bigcache parameters
func New() *Client {
	bc, _ := bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Minute))
	return &Client{bc: bc}
}

// NewWithClient returns a new driver client with the given bigcache.BigCache configuration
func NewWithClient(bc *bigcache.BigCache) *Client {
	return &Client{bc: bc}
}

// Get implements the "cache.Cache".Get() interface
func (c *Client) Get(key string, dstVal interface{}) error {
	b, err := c.bc.Get(key)
	if err != nil {
		return err
	}
	if err := Codec.Unmarshal(b, dstVal); err != nil {
		return err
	}
	return nil
}

// Set implements the "cache.Cache".Set() interface. Bigcache doesn't support
// expirations based on time, nor does it support deleting keys, so anything other
// than a 0 exp will result in an error.
func (c *Client) Set(key string, val interface{}, exp time.Duration) error {
	if exp > cache.NeverExpires {
		return cache.ErrUnsupportedAction
	}
	b, err := Codec.Marshal(val)
	if err != nil {
		return err
	}
	return c.bc.Set(key, b)
}

// Exists implements the "cache.Cache".Exists() interface
func (c *Client) Exists(key string) bool {
	var v interface{}
	return c.Get(key, v) == nil
}

// Del implements the "cache.Cache".Del() interface
func (c *Client) Del(key string) error {
	return cache.ErrUnsupportedAction
}

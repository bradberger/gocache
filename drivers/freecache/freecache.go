package freecache

import (
	"time"

	"github.com/bradberger/gocache/cache"
	"github.com/bradberger/gocache/codec"
	"github.com/coocood/freecache"
)

var (
	// Codec is the codec used to marshal/unmarshal interfaces into the byte slices required by the Bigcache client
	Codec = codec.Gob

	_ cache.Cache = (*Client)(nil)
)

func New(size int) *Client {
	return &Client{fc: freecache.NewCache(size)}
}

type Client struct {
	fc *freecache.Cache
}

func (c *Client) Get(key string, dstVal interface{}) error {
	b, err := c.fc.Get([]byte(key))
	if err != nil {
		if err == freecache.ErrNotFound {
			err = cache.ErrNotFound
		}
		return err
	}
	if err := Codec.Unmarshal(b, dstVal); err != nil {
		return err
	}
	return nil
}

func (c *Client) Set(key string, val interface{}, exp time.Duration) error {
	b, err := Codec.Marshal(val)
	if err != nil {
		return err
	}
	return c.fc.Set([]byte(key), b, int(exp/time.Second))
}

func (c *Client) Exists(key string) bool {
	var v interface{}
	return c.Get(key, v) == nil
}

func (c *Client) Del(key string) error {
	return cache.ErrUnsupportedAction
}

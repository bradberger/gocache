package arc

import (
	"time"

	"github.com/bradberger/gocache/cache"
	"github.com/hashicorp/golang-lru"
)

var (
	_ cache.Cache = (*Client)(nil)
)

type Client struct {
	c *lru.ARCCache
}

func New(size int) (*Client, error) {
	c, err := lru.NewARC(size)
	if err != nil {
		return nil, err
	}
	return &Client{c: c}, nil
}

func (c *Client) Set(key string, val interface{}, exp time.Duration) error {
	if exp != cache.NeverExpires {
		return cache.ErrUnsupportedAction
	}
	c.c.Add(key, val)
	return nil
}

func (c *Client) Get(key string, dstVal interface{}) error {
	v, ok := c.c.Get(key)
	if !ok {
		return cache.ErrNotFound
	}
	return cache.Copy(v, dstVal)
}

func (c *Client) Exists(key string) bool {
	return c.c.Contains(key)
}

func (c *Client) Del(key string) error {
	c.c.Remove(key)
	return nil
}

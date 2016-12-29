// Package redis implements a driver for github.com/garyburd/redigo/redis
package redis

import (
	"fmt"
	"time"

	"github.com/bradberger/gocache/cache"
	"github.com/bradberger/gocache/codec"
	"github.com/garyburd/redigo/redis"
)

var (
	// Codec is the codec used to marshal/unmarshal interfaces into the byte slices required by the Diskv client
	Codec codec.Codec

	// ensure struct implements the cache.KeyStore interface
	_ cache.Cache = (*Client)(nil)
)

func init() {
	Codec = codec.Gob
}

type Client struct {
	pool *redis.Pool
}

// DefaultClient provides a local Redis client with 10 maximum concurrent connections
var DefaultClient = New(":6379", 10)

// New returns a Redist client to the given server with maxConn maximum concurrent connections
func New(addr string, maxConn int) *Client {
	return &Client{
		pool: redis.NewPool(func() (redis.Conn, error) {
			return redis.Dial("tcp", addr)
		}, maxConn),
	}
}

// Get implements the "cache.KeyStore".Get() interface
func (c *Client) Get(key string, dstVal interface{}) error {
	conn := c.Pool().Get()
	defer conn.Close()
	b, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			err = cache.ErrCacheMiss
		}
		return err
	}
	return Codec.Unmarshal(b, dstVal)
}

// Set implements the "cache.KeyStore".Set() interface
func (c *Client) Set(key string, val interface{}, exp time.Duration) error {
	b, err := Codec.Marshal(val)
	if err != nil {
		return err
	}
	conn := c.Pool().Get()
	defer conn.Close()

	if exp > 0 {
		_, err = conn.Do("SET", key, b, "PX", fmt.Sprintf("%d", int64(exp.Seconds()*1000)))
	} else {
		_, err = conn.Do("SET", key, b)
	}
	return err
}

// Del implements the "cache.KeyStore".Del() interface
func (c *Client) Del(key string) error {
	conn := c.Pool().Get()
	defer conn.Close()
	_, err := conn.Do("DEL", key)
	return err
}

// Exists implements the "cache.Cache".Exists() interface
func (c *Client) Exists(key string) bool {
	conn := c.Pool().Get()
	defer conn.Close()
	b, err := redis.Bytes(conn.Do("GET", key))
	return err == nil || b != nil
}

// Pool returns the underlying redis.Pool
func (c *Client) Pool() *redis.Pool {
	return c.pool
}

// Close returns the underlying Redis pool
func (c *Client) Close() error {
	return c.Pool().Close()
}

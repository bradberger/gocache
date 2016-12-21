package memcache

import (
	"time"

	"github.com/bradberger/gocache/cache"
	"github.com/bradberger/gocache/codec"
	"github.com/bradfitz/gomemcache/memcache"
)

var (
	// Codec is the codec used to marshal/unmarshal interfaces into the byte slices required by the Memcache client
	Codec codec.Codec

	// ensure LRU struct implements the cache.KeyStore interface
	_ cache.Cache = (*Memcache)(nil)
)

func init() {
	Codec = codec.Gob
}

// Memcache is a struct which implements cache.Cache interface and connects to a Memcache server
type Memcache struct {
	mc *memcache.Client
}

// New initializes a memcache Client which connects to the given servers
func New(servers ...string) *Memcache {
	return &Memcache{mc: memcache.New(servers...)}
}

// Set sets the given key/value pair
func (m *Memcache) Set(key string, value interface{}, exp time.Duration) error {
	b, err := Codec.Marshal(value)
	if err != nil {
		return err
	}
	return m.Client().Set(&memcache.Item{Key: key, Value: b, Expiration: int32(exp / time.Second)})
}

// Get returns the value for the given key
func (m *Memcache) Get(key string, dstVal interface{}) error {
	item, err := m.Client().Get(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			err = cache.ErrCacheMiss
		}
		return err
	}
	return Codec.Unmarshal(item.Value, dstVal)
}

// Del deletes the given key
func (m *Memcache) Del(key string) error {
	return m.Client().Delete(key)
}

// Exists implements the "cache.Cache".Exists() interface
func (m *Memcache) Exists(key string) bool {
	var v interface{}
	return m.Get(key, v) == nil
}

// Add writes the given item, if no value already exists for its key. ErrNotStored is returned if that condition is not met.
func (m *Memcache) Add(key string, val interface{}, exp time.Duration) error {
	b, err := Codec.Marshal(val)
	if err != nil {
		return err
	}
	return m.Client().Add(&memcache.Item{Key: key, Value: b, Expiration: int32(exp / time.Second)})
}

// Replace writes the given item, but only if the server *does* already hold data for this key
func (m *Memcache) Replace(key string, val interface{}, exp time.Duration) error {
	b, err := Codec.Marshal(val)
	if err != nil {
		return err
	}
	return m.Client().Replace(&memcache.Item{Key: key, Value: b, Expiration: int32(exp / time.Second)})
}

// Touch updates a key's expiration
func (m *Memcache) Touch(key string, exp time.Duration) error {
	return m.Client().Touch(key, int32(exp/time.Second))
}

// Increment increases the key's value by delta
func (m *Memcache) Increment(key string, delta uint64) error {
	_, err := m.Client().Increment(key, delta)
	return err
}

// Decrement decreases the key's value by delta
func (m *Memcache) Decrement(key string, delta uint64) error {
	_, err := m.Client().Decrement(key, delta)
	return err
}

// Flush removes all keys from the cache
func (m *Memcache) Flush() error {
	return m.Client().FlushAll()
}

// Client returns the underlying Memcache client to enable usage of it's more powerful feature-set
func (m *Memcache) Client() *memcache.Client {
	return m.mc
}

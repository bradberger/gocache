package lru

import (
	"fmt"
	"time"

	"stathat.com/c/consistent"
)

// ShardedLRU is a LRU with keys split across shards for better performance
type ShardedLRU struct {
	shards map[string]*LRU
	ch     *consistent.Consistent
}

// New returns a new ShardedLRU with a up to a max of one Gigabyte of memory usage
func New() *ShardedLRU {
	return NewSharded(3<<30, Gigabyte, 256)
}

// NewSharded returns a new sharded LRU cache with a maximum in-memory size of maxSize,
// and total maxEntries spread over shardCt
func NewSharded(maxSize uint64, maxEntries uint64, shardCt uint64) *ShardedLRU {

	l := int(shardCt)
	shardSize := maxSize / shardCt
	shardEntries := maxEntries / shardCt

	var k string
	c := ShardedLRU{make(map[string]*LRU), consistent.New()}
	for i := 0; i < l; i++ {
		k = fmt.Sprintf("shard-%2d", i)
		c.ch.Add(k)
		c.shards[k] = NewBasic(shardSize, int(shardEntries))
	}

	return &c
}

func (c *ShardedLRU) find(key string) *LRU {
	nm, _ := c.ch.Get(key)
	return c.shards[nm]
}

// CompareAndSwap writes the given item that was previously returned by Get,
// if the value was neither modified or evicted between the Get and the
// CompareAndSwap calls. The item's Key should not change between calls but
// all other item fields may differ. ErrCASConflict is returned if the value
//  was modified in between the calls. ErrNotStored is returned if the value
// was evicted in between the calls.
func (c *ShardedLRU) CompareAndSwap(key string, value interface{}, exp time.Duration, cas uint64) (err error) {
	return c.find(key).CompareAndSwap(key, value, exp, cas)
}

// SetItem sets the item with the given parameters. Use it if you need to specifiy the cas id
// for future CompareAndSwap actions. Otherwise it's easier to just use Set() instead.
func (c *ShardedLRU) SetItem(key string, value interface{}, exp time.Duration, cas uint64) (err error) {
	return c.find(key).SetItem(key, value, exp, cas)
}

// Set sets the key/value pair
func (c *ShardedLRU) Set(key string, value interface{}, exp time.Duration) (err error) {
	return c.find(key).Set(key, value, exp)
}

// GetItem is similar to Get() only it returns the items unique CAS identifier
func (c *ShardedLRU) GetItem(key string, dstVal interface{}) (cas uint64, err error) {
	return c.find(key).GetItem(key, dstVal)
}

// Get fetches the key into the dstVal. It returns an error if the value doesn't exist.
func (c *ShardedLRU) Get(key string, dstVal interface{}) (err error) {
	return c.find(key).Get(key, dstVal)
}

// Del removes the key from the cache
func (c *ShardedLRU) Del(key string) (err error) {
	return c.find(key).Del(key)
}

// Exists returns true if the given key exists
func (c *ShardedLRU) Exists(key string) bool {
	return c.find(key).Exists(key)
}

// Touch updates the expiry for the given key
func (c *ShardedLRU) Touch(key string, exp time.Duration) (err error) {
	return c.find(key).Touch(key, exp)
}

// Add adds the value to cache, but only if the key doesn't already exist.
func (c *ShardedLRU) Add(key string, value interface{}, exp time.Duration) (err error) {
	return c.find(key).Add(key, value, exp)
}

// Replace replaces the current value for the key with the new value, but only
// if the key already exists. The expiration remains the same.
func (c *ShardedLRU) Replace(key string, value interface{}) (err error) {
	return c.find(key).Replace(key, value)
}

// Increment increases the key's value by delta. The exipration remains the same.
// The existing value must be a uint64 or a pointer to a uint64
func (c *ShardedLRU) Increment(key string, delta uint64) (uint64, error) {
	return c.find(key).Increment(key, delta)
}

// Decrement decreases the key's value by delta. The expiration remains the same.
// The underlying value to increment must be a uint64 or pointer to a uint64.
func (c *ShardedLRU) Decrement(key string, delta uint64) (uint64, error) {
	return c.find(key).Decrement(key, delta)
}

// Append appends the value to the current key value. The value to append
// must be a string, []byte or pointer to one of those types.
func (c *ShardedLRU) Append(key string, value interface{}) (err error) {
	return c.find(key).Append(key, value)
}

// Prepend prepends the value to the current key value. The value to prepend
// must be a string, []byte or pointer to one of those types.
func (c *ShardedLRU) Prepend(key string, value interface{}) (err error) {
	return c.find(key).Prepend(key, value)
}

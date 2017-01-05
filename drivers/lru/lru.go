// Package lru implements a least recently used in-memory cache. It used
// a doubly linked list and purges items if the number of items in the cache
// exceeds the max, the total size of the cache in memory exceeds the given
// size, or the item expires.
package lru

import (
	"container/list"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bradberger/gocache/cache"
)

// Sizes and defaults for the cache
const (
	Byte     uint64 = 1
	Kilobyte        = Byte << 10
	Megabyte        = Kilobyte << 10
	Gigabyte        = Megabyte << 10
	Terabyte        = Gigabyte << 10
)

var (
	// ensure LRU struct implements the cache.KeyStore interface
	_ cache.Cache = (*LRU)(nil)

	// TickerDuration is the duration between housekeeping for deleting values
	// if the total cache size is too large, or the max entries are too many,
	// or items are expired. By default it's one minute, but that definitely
	// should be tuned according to your use-case scenario. Expired items will
	// never be returned with Get() regardless of this setting. This is more
	// about memory management and "garbage collection" pauses more than anything
	// else.
	TickerDuration = 1 * time.Minute

	// ErrCannotSetValue is returned when an previously set cache value pointer cannot be set.
	ErrCannotSetValue = errors.New("cannot set value of interface")

	// ErrCannotAssignValue is returned when a previously set cache value pointer cannot be
	// updated because the new value's type cannot be assigned to the previous value's type.
	ErrCannotAssignValue = errors.New("cannot assign value")

	getEntryPool = sync.Pool{
		New: func() interface{} {
			return &entry{}
		},
	}
)

// LRU is a least recently used in-memory cache implementation
type LRU struct {
	// MaxEntries are the max number of cache entries. 0 is unlimited
	MaxEntries int
	// MaxSize is the max size, in bytes, of the total cache store in memory
	MaxSize uint64

	ll    *list.List
	cache map[string]*list.Element
	size  uint64

	sync.RWMutex
}

type entry struct {
	key   string
	value interface{}
	exp   time.Time
}

// New creates a new LRU cache with the given number of max entries
func New(maxSize uint64, maxEntries int) *LRU {
	c := &LRU{
		MaxEntries: maxEntries,
		MaxSize:    maxSize,
		cache:      make(map[string]*list.Element, maxEntries),
		ll:         list.New(),
	}
	return c
}

// Set sets the key/value pair
func (c *LRU) Set(key string, value interface{}, exp time.Duration) (err error) {

	// Dereference pointers to optimize GC cycles
	if el := reflect.ValueOf(value); el.Kind() == reflect.Ptr {
		value = el.Elem().Interface()
	}

	ent := getEntryPool.Get().(*entry)
	ent.key = key
	ent.value = value
	if exp != cache.NeverExpires {
		ent.exp = time.Now().Add(exp)
		go func(key string, exp time.Duration) {
			time.Sleep(exp)
			c.Lock()
			if ee, ok := c.cache[key]; ok && ee.Value.(*entry).exp.Before(time.Now()) {
				c.delete(key)
			}
			c.Unlock()
		}(key, exp)
	}

	atomic.AddUint64(&c.size, uint64(reflect.TypeOf(ent).Size()))
	c.Lock()
	if ee, ok := c.cache[key]; ok {
		atomic.AddUint64(&c.size, ^(uint64(reflect.TypeOf(ee.Value).Size()) - 1))
		ee.Value = ent
		c.ll.MoveToFront(ee)
	} else {
		ele := c.ll.PushFront(ent)
		c.cache[key] = ele
		if c.ll.Len() > c.MaxEntries {
			c.delete(c.ll.Back().Value.(*entry).key)
		}
	}
	for c.size > c.MaxSize {
		if b := c.ll.Back(); b != nil {
			c.delete(b.Value.(*entry).key)
		}
	}
	c.Unlock()

	return nil
}

// Get fetches the key into the dstVal. It returns an error if the value doesn't exist.
func (c *LRU) Get(key string, dstVal interface{}) (err error) {

	// Lock for writing
	c.Lock()
	defer c.Unlock()

	ele, hit := c.cache[key]
	if !hit {
		return cache.ErrCacheMiss
	}

	// Make sure the key is not expired. This allows us to relax the housekeeping
	// of expired values a bit, which should reduce "Gc" style pauses.
	if exp := ele.Value.(*entry).exp; !exp.IsZero() && exp.Before(time.Now()) {
		return cache.ErrCacheMiss
	}

	// Move element to front
	c.ll.MoveToFront(ele)
	return cache.Copy(ele.Value.(*entry).value, dstVal)
}

func (c *LRU) delete(key string) {
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

// Del removes the key from the cache
func (c *LRU) Del(key string) (err error) {
	c.Lock()
	c.delete(key)
	c.Unlock()
	return
}

// Exists returns true if the given key exists
func (c *LRU) Exists(key string) (exists bool) {
	c.RLock()
	_, exists = c.cache[key]
	c.RUnlock()
	return
}

// Len returns the number of items in the cache.
func (c *LRU) Len() int {
	c.RLock()
	l := c.ll.Len()
	c.RUnlock()
	return l
}

func (c *LRU) removeElement(e *list.Element) {
	c.removeElements([]*list.Element{e})
}

func (c *LRU) removeElements(e []*list.Element) {
	for i := range e {
		if prev := c.ll.Remove(e[i]); prev != nil {
			kv := prev.(*entry)
			atomic.AddUint64(&c.size, ^uint64(uint64(reflect.TypeOf(kv).Size())-1))
			delete(c.cache, kv.key)
			getEntryPool.Put(kv)
		}
	}
}

// Touch updates the expiry for the given key
func (c *LRU) Touch(key string, exp time.Duration) error {

	c.Lock()
	defer c.Unlock()
	ee, ok := c.cache[key]
	if !ok {
		return cache.ErrCacheMiss
	}

	ent := ee.Value.(*entry)
	switch {
	case exp == cache.NeverExpires:
		ent.exp = time.Time{}
	default:
		ent.exp = time.Now().Add(exp)
	}
	c.ll.MoveToFront(ee)
	return nil
}

// Add adds the value to cache, but only if the key doesn't already exist.
func (c *LRU) Add(key string, value interface{}, exp time.Duration) error {
	if c.Exists(key) {
		return cache.ErrNotStored
	}
	return c.Set(key, value, exp)
}

// Replace replaces the current value for the key with the new value, but only
// if the key already exists. The expiration remains the same.
func (c *LRU) Replace(key string, value interface{}) error {
	c.Lock()
	defer c.Unlock()
	e, exists := c.cache[key]
	if !exists {
		return cache.ErrNotStored
	}
	e.Value.(*entry).value = value
	c.ll.MoveToFront(e)
	return nil
}

// Increment increases the key's value by delta. The exipration remains the same.
// The existing value must be a uint64 or a pointer to a uint64
func (c *LRU) Increment(key string, delta uint64) (uint64, error) {

	c.RLock()
	e, exists := c.cache[key]
	c.RUnlock()
	if !exists {
		return 0, cache.ErrNotStored
	}

	v, ok := e.Value.(*entry).value.(uint64)
	if !ok {
		return 0, ErrCannotAssignValue
	}

	v += delta

	c.Lock()
	e.Value.(*entry).value = v
	c.ll.MoveToFront(e)
	c.Unlock()

	return v, nil
}

// Decrement decreases the key's value by delta. The expiration remains the same.
// The underlying value to increment must be a uint64 or pointer to a uint64.
func (c *LRU) Decrement(key string, delta uint64) (uint64, error) {
	c.RLock()
	e, exists := c.cache[key]
	c.RUnlock()
	if !exists {
		return 0, cache.ErrNotStored
	}
	v, ok := e.Value.(*entry).value.(uint64)
	if !ok || delta > v {
		return 0, cache.ErrCannotAssignValue
	}
	v -= delta
	c.Lock()
	e.Value.(*entry).value = v
	c.ll.MoveToFront(e)
	c.Unlock()

	return v, nil
}

// Prepend prepends the value to the current key value. The value to prepend
// must be a string, []byte or pointer to one of those types.
func (c *LRU) Prepend(key string, value interface{}) error {

	var toPrepend []byte
	switch value.(type) {
	case []byte:
		toPrepend = value.([]byte)
	case *[]byte:
		toPrepend = *value.(*[]byte)
	case string:
		toPrepend = []byte(value.(string))
	case *string:
		toPrepend = []byte(*value.(*string))
	default:
		return ErrCannotAssignValue
	}

	c.Lock()
	defer c.Unlock()
	e, exists := c.cache[key]
	if !exists {
		return cache.ErrNotStored
	}
	ee := e.Value.(*entry)
	switch ee.value.(type) {
	case []byte:
		ee.value = append(toPrepend, ee.value.([]byte)...)
	case string:
		ee.value = string(toPrepend) + ee.value.(string)
	default:
		return ErrCannotAssignValue
	}
	c.ll.MoveToFront(e)
	return nil
}

// Append appends the value to the current key value. The value to append
// must be a string, []byte or pointer to one of those types.
func (c *LRU) Append(key string, value interface{}) error {

	var toAppend []byte
	switch value.(type) {
	case []byte:
		toAppend = value.([]byte)
	case *[]byte:
		toAppend = *value.(*[]byte)
	case string:
		toAppend = []byte(value.(string))
	case *string:
		toAppend = []byte(*value.(*string))
	default:
		return ErrCannotAssignValue
	}

	c.Lock()
	defer c.Unlock()
	e, exists := c.cache[key]
	if !exists {
		return cache.ErrNotStored
	}
	ee := e.Value.(*entry)
	switch ee.value.(type) {
	case []byte:
		ee.value = append(ee.value.([]byte), toAppend...)
	case string:
		ee.value = ee.value.(string) + string(toAppend)
	default:
		return ErrCannotAssignValue
	}
	c.ll.MoveToFront(e)
	return nil
}

// Package memory implements a very basic, fully functional cache.Cache interface
// with in-memory caching and key expiration. Expired keys are purged immediately.
// There are no controls over the size of the cache, so for long-standing caches,
// the lru package is a better option as it manages the cache in-memory size.
package memory

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/bradberger/gocache/cache"
)

var (
	// ensure Memory struct implements the cache.Cache interface
	_ cache.Cache = (*Memory)(nil)

	// ErrCannotAssignValue is returned when the dst value cannot be assigned
	ErrCannotAssignValue = errors.New("cannot assign value to dst")
)

// Memory implements a very basic, fully functional cache.Cache interface with in-memory caching and key expiration.
type Memory struct {
	m   map[string]interface{}
	exp map[string]time.Time
	sync.RWMutex
}

// New returns a new Memory example cache
func New() *Memory {
	return &Memory{m: make(map[string]interface{}), exp: make(map[string]time.Time)}
}

// Get implements the "cache.Cache".Get() interface
func (c *Memory) Get(key string, dstVal interface{}) error {
	dstEl := reflect.ValueOf(dstVal)
	if dstEl.Kind() == reflect.Ptr {
		dstEl = dstEl.Elem()
	}
	if !dstEl.CanSet() {
		return cache.ErrInvalidDstVal
	}
	c.RLock()
	val, found := c.m[key]
	c.RUnlock()
	if !found {
		return cache.ErrCacheMiss
	}

	curEl := reflect.ValueOf(val)
	if curEl.Kind() == reflect.Ptr {
		curEl = curEl.Elem()
	}
	if !curEl.Type().AssignableTo(dstEl.Type()) {
		return ErrCannotAssignValue
	}
	dstEl.Set(curEl)
	return nil
}

// Set implements the "cache.Cache".Set() interface
func (c *Memory) Set(key string, val interface{}, exp time.Duration) error {
	c.Lock()
	defer c.Unlock()

	cur, ok := c.m[key]
	for {

		if !ok {
			c.m[key] = val
			break
		}

		// If the current element is a pointer, then we must check to see that
		// the new value being set is compatible, otherwise panics could ensue...
		curEl := reflect.ValueOf(cur)
		if curEl.Kind() != reflect.Ptr {
			c.m[key] = val
			break
		}

		// So here we check to make sure that it's assignable, and if so then set it.
		newEl := reflect.ValueOf(val)
		curEl = curEl.Elem()
		if !curEl.CanSet() || !curEl.Type().AssignableTo(newEl.Type()) {
			return ErrCannotAssignValue
		}
		curEl.Set(newEl)
		break
	}

	// If no expiration, then just return.
	if exp < 1 {
		return nil
	}

	// Otherwise, schedule a check which will try to expire the key.
	// If it's been updated in the meantime with newer expiration, then
	// the following check will fail and it won't be deleted. Otherwise it
	// will be deleted.
	c.exp[key] = time.Now().Add(exp)
	go func(key string, delay time.Duration) {
		select {
		case <-time.After(delay):
			if c.expires(key).Before(time.Now()) {
				c.Del(key)
			}
		}
	}(key, exp)
	return nil
}

func (c *Memory) expires(key string) (exp time.Time) {
	c.RLock()
	exp, _ = c.exp[strings.TrimSpace(key)]
	c.RUnlock()
	return
}

// Del implements the "cache.Cache".Del() interface
func (c *Memory) Del(key string) error {
	c.Lock()
	delete(c.m, key)
	delete(c.exp, key)
	c.Unlock()
	return nil
}

// Exists implements the "cache.Cache".Exists() interface
func (c *Memory) Exists(key string) bool {
	c.RLock()
	_, exists := c.m[key]
	c.RUnlock()
	return exists
}

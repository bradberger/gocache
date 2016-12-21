package cache

import (
	"errors"
	"fmt"
	"time"
)

var (
	// ErrCacheMiss is returned when the key isn't in the cache
	ErrCacheMiss = errors.New("cache miss")
	// ErrInvalidDstVal is returned with the dstVal of the Get() func cannot be set. It's probably because the
	// dstVal is not a pointer.
	ErrInvalidDstVal = errors.New("cannot set dst value")
	// ErrInvalidDataFormat is returned when the data retrieved from a storage engine is not in the expected format
	ErrInvalidDataFormat = errors.New("Invalid data format")
	// ErrKeyExists indicates that the item you are trying to store with
	// a "cas" command has been modified since you last fetched it.
	ErrKeyExists = errors.New("key already exists")
	// ErrNotStored indicate the data was not stored, but not
	// because of an error. This normally means that the
	// condition for an "add" or a "replace" command wasn't met.
	ErrNotStored = errors.New("not stored")
	// ErrNotFound indicates that the item you are trying to store
	// with a "cas" command did not exist.
	ErrNotFound = errors.New("not found")
)

// Cache defines a cache with key expiration. Implementations of this interface
// are expected to be thread safe.
type Cache interface {
	Set(key string, value interface{}, expiration time.Duration) error
	Get(key string, dstVal interface{}) error
	Exists(key string) bool
	Del(key string) error
}

// Cachestore defines a cache interface which supports exporting all it's keys and also
// transferring all it's data to another Cache.
type Cachestore interface {
	Cache
	KeyList
	Transfer(Cache) error
}

// KeyProvider is an interface which can describe it's own key. It's used for getting/setting
// key/value pairs without a directly supplied key string. Instead, the supplied interface
// can announce it's own key, and that's used in getting/setting.
type KeyProvider interface {
	Key() string
}

// Store defines a permanent key/value store. The storage mechanism should be
// permanent, not volitile.
type Store interface {
	Set(key string, value interface{}) error
	Get(key string, dstVal interface{}) error
	Del(key string) error
}

// Flush defines an interface which store can clear all it's key/value pairs at once.
type Flush interface {
	Flush() error
}

// Touch defines an interface for caches which can touch item's expiration time.
type Touch interface {
	Touch(key string, exp time.Duration) error
}

// Increment defines an interface for caches which can increment a key's value
type Increment interface {
	Increment(key string, delta uint64) (uint64, error)
}

// Decrement defines an interface for caches which can decrement a key's value
type Decrement interface {
	Decrement(key string, delta uint64) (uint64, error)
}

// KeyList defines an interface for announcing all keys currently set
type KeyList interface {
	Keys() []string
}

// Append defines an interface which appends the value to the current value for the
// key. If the key doesn't already exist, it must return a ErrCacheMiss error
type Append interface {
	Append(key string, value interface{}) error
}

// Prepend defines an interface which prepends the value to the current value for the
// key. If the key doesn't already exist, it must return a ErrCacheMiss error
type Prepend interface {
	Prepend(key string, value interface{}) error
}

// Replace defines an interface which replaces the value to the current value for the
// key. If the key doesn't already exist, it must return a ErrNotStored error
type Replace interface {
	Replace(key string, value interface{}) error
}

// Add defines an interface which adds the key/value to the cache but only if
// the key doesn't already exist in the cache. If it does exist, it must return
// an ErrNotStored error
type Add interface {
	Add(key string, value interface{}, exp time.Duration) error
}

// Key turns Stringer funcs, byte slices, pointers to strings, etc., into string keys
func Key(key interface{}) string {
	if kp, ok := key.(KeyProvider); ok {
		return kp.Key()
	}
	switch key.(type) {
	case string:
		return key.(string)
	case *string:
		return *key.(*string)
	case []byte:
		return string(key.([]byte))
	case *[]byte:
		return string(*key.(*[]byte))
	default:
		return fmt.Sprintf("%#v", key)
	}
}

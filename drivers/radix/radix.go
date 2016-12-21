package radix

import (
	"reflect"
	"time"

	goradix "github.com/armon/go-radix"
	"github.com/bradberger/gocache/cache"
)

var (
	_ cache.Cache = (*Tree)(nil)
)

// Tree implements a cache store with a radix tree
type Tree struct {
	tree *goradix.Tree
}

// New returns an initialized tree
func New() *Tree {
	return &Tree{tree: goradix.New()}
}

// Get implements the "cache.KeyStore".Get() interface
func (t *Tree) Get(key string, dstVal interface{}) error {
	val, found := t.Tree().Get(key)
	if !found {
		return cache.ErrCacheMiss
	}
	el := reflect.ValueOf(dstVal)
	if el.Kind() == reflect.Ptr {
		el = el.Elem()
	}
	if !el.CanSet() {
		return cache.ErrInvalidDstVal
	}
	el.Set(reflect.ValueOf(val))
	return nil
}

// Set implements the "cache.KeyStore".Set() interface
func (t *Tree) Set(key string, val interface{}, exp time.Duration) error {
	t.Tree().Insert(key, val)
	if exp > 0 {
		go func(key string, exp time.Duration) {
			select {
			case <-time.After(exp):
				t.Del(key)
			}
		}(key, exp)
	}
	return nil
}

// Del implements the "cache.KeyStore".Del() interface
func (t *Tree) Del(key string) error {
	t.Tree().Delete(key)
	return nil
}

// Exists returns true if the key exists
func (t *Tree) Exists(key string) bool {
	_, found := t.Tree().Get(key)
	return found
}

// Tree returns the underlying go-radix tree for a more complete featureset
func (t *Tree) Tree() *goradix.Tree {
	return t.tree
}

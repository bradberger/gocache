package btree

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/bradberger/gocache/cache"
	"github.com/google/btree"

	"stathat.com/c/consistent"
)

var (
	_ cache.Cache = (*Tree)(nil)
	_ btree.Item  = (*item)(nil)
)

type item struct {
	key   string
	value interface{}
}

func (i *item) Less(than btree.Item) bool {
	if than == nil {
		return true
	}
	return strings.Compare(i.key, than.(*item).key) == -1
}

type Container struct {
	shards map[string]*Tree
	ch     *consistent.Consistent
}

func NewContainer(numShards int) *Container {
	c := &Container{shards: make(map[string]*Tree, numShards), ch: consistent.New()}
	c.ch.NumberOfReplicas = 1
	for i := 0; i < numShards; i++ {
		k := fmt.Sprintf("shard-%d", i)
		c.shards[k] = New()
		c.ch.Add(k)
	}
	return c
}

func (c *Container) tree(key string) (*Tree, error) {
	name, err := c.ch.Get(key)
	if err != nil {
		return nil, err
	}
	return c.shards[name], nil
}

func (c *Container) Get(key string, dstVal interface{}) (err error) {
	t, err := c.tree(key)
	if err != nil {
		return err
	}
	return t.Get(key, dstVal)
}

func (c *Container) Set(key string, val interface{}, exp time.Duration) error {
	t, err := c.tree(key)
	if err != nil {
		return err
	}
	return t.Set(key, val, exp)
}

func (c *Container) Del(key string) error {
	t, err := c.tree(key)
	if err != nil {
		return err
	}
	return t.Del(key)
}

func (c *Container) Exists(key string) bool {
	t, err := c.tree(key)
	if err != nil {
		return false
	}
	return t.Exists(key)
}

// Tree is a wrapper for github.com/google/btree which adds expiration times and parallel thread write safetey
type Tree struct {
	tree *btree.BTree

	exp map[string]time.Time

	sync.RWMutex
}

// New returns a newly initialized btree based cache with a default degree of 2
func New() *Tree {
	return &Tree{tree: btree.New(2), exp: make(map[string]time.Time)}
}

// Get implements the "cache.Cache".Get interface
func (t *Tree) Get(key string, dstVal interface{}) error {
	i, ok := t.tree.Get(&item{key: key}).(*item)
	if !ok || i == nil || i.value == nil {
		return cache.ErrNotFound
	}
	return cache.Copy(i.value, dstVal)
}

// Set implements the "cache.Cache".Set interface
func (t *Tree) Set(key string, val interface{}, exp time.Duration) error {
	if reflect.ValueOf(val).Kind() == reflect.Ptr {
		val = reflect.ValueOf(val).Elem().Interface()
	}
	t.Lock()
	t.tree.ReplaceOrInsert(&item{key: key, value: val})
	if exp > 0 {
		t.exp[key] = time.Now().Add(exp)
		go t.expire(key, exp)
	}
	t.Unlock()
	return nil
}

func (t *Tree) expire(key string, delay time.Duration) {
	select {
	case <-time.After(delay):
		t.RLock()
		exp, _ := t.exp[key]
		t.RUnlock()
		if exp.Before(time.Now()) {
			t.Del(key)
		}
	}
}

// Del implements the "cache.Cache".Del interface
func (t *Tree) Del(key string) error {
	t.Lock()
	t.tree.Delete(&item{key: key})
	delete(t.exp, key)
	t.Unlock()
	return nil
}

// Exists implements the "cache.Cache".Exists interface
func (t *Tree) Exists(key string) bool {
	return t.tree.Has(&item{key: key})
}

// Tree returns the underlying btree
func (t *Tree) Tree() *btree.BTree {
	return t.tree
}

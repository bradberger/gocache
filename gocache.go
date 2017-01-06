// Package gocache implements a set of drivers and a common interface for working with different cache systems
package gocache

import (
	"errors"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/bradberger/gocache/cache"
	"stathat.com/c/consistent"
)

const (
	// ReplicateAsync indicates that replication will be done asyncronously.
	// Set commands will return without error as soon as at least one node has
	// the value
	ReplicateAsync ReplicationMethod = iota

	// ReplicateSync indicates that replication will be done syncronously.
	// Set commands will return without error only if all nodes return without error
	ReplicateSync = iota

	// RetrieveSync indicates that get commands(get, exists) will be send to one
	// node at a time starting with the closest node, and the first result will be used
	RetrieveSync RetrievalMethod = iota

	// RetrieveAsync indicates that get commands(get, exists) will be send in parallel
	// to the given nodes, and the first result will be used
	RetrieveAsync = iota
)

var (
	// ErrGetTimeout indicates a retrieval command timed out
	ErrGetTimeout = errors.New("get timeout")
	// ErrSetTimeout indicates a storage command timed out
	ErrSetTimeout             = errors.New("set timeout")
	_             cache.Cache = (*Client)(nil)
)

// ReplicationMethod determines whether replication takes place asyncronously or syncronously.
// Use ReplicateAsync for asyncronous replication, ReplicateSync for syncronous replication.
type ReplicationMethod int

// RetrievalMethod determines whether get commands (get, exists) are sent to all nodes
// simultaneously or are sent one-by-one.
type RetrievalMethod int

// Client is a cache client with built in replication to any number of different caches.
// This allows replication and syncronization across various caches using the set of drivers
// available as subpackages, including Memcached, Redis, in-memory caches, and more.
type Client struct {
	nodes map[string]cache.Cache
	ch    *consistent.Consistent

	replicateNodeCt int
	replicateMethod ReplicationMethod
	retrievalMethod RetrievalMethod

	timeoutSet time.Duration
	timeoutGet time.Duration

	sync.RWMutex
}

// New returns a new initialized cache Client with no nodes.
func New() *Client {
	return &Client{
		nodes:           make(map[string]cache.Cache, 0),
		ch:              consistent.New(),
		replicateMethod: ReplicateAsync,
		timeoutSet:      time.Millisecond * 200,
		timeoutGet:      time.Millisecond * 200,
	}
}

// AddNode adds a cache node with the given name, but only if it doesn't already exist
func (c *Client) AddNode(name string, node cache.Cache) error {
	c.RLock()
	_, exists := c.nodes[name]
	c.RUnlock()
	if exists {
		return errors.New("node already exists")
	}
	return c.SetNode(name, node)
}

// SetNode sets the cache node with the given name, regardless of whether it already exists or not
func (c *Client) SetNode(name string, node cache.Cache) error {
	if node == nil {
		return errors.New("cache node is nil")
	}
	c.Lock()
	c.nodes[name] = node
	c.ch.Add(name)
	c.Unlock()
	return nil
}

// GetTimeout updates the get command timeout duration
func (c *Client) GetTimeout(dur time.Duration) {
	c.Lock()
	c.timeoutGet = dur
	c.Unlock()
}

// SetTimeout updates the set command timeout duration
func (c *Client) SetTimeout(dur time.Duration) {
	c.Lock()
	c.timeoutSet = dur
	c.Unlock()
}

// ReplaceNode adds a cache node with the given name, but only if it already exists
func (c *Client) ReplaceNode(name string, node cache.Cache) error {
	c.RLock()
	_, exists := c.nodes[name]
	c.RUnlock()
	if !exists {
		return errors.New("node does not exist")
	}
	return c.SetNode(name, node)
}

// RemoveNode removes a node with the given name from the node list
func (c *Client) RemoveNode(name string) error {
	c.Lock()
	defer c.Unlock()
	delete(c.nodes, name)
	c.ch.Remove(name)
	return nil
}

// SetReplicateMethod sets the replication method
func (c *Client) SetReplicateMethod(m ReplicationMethod) {
	c.Lock()
	c.replicateMethod = m
	c.Unlock()
}

// SetRetrievalMethod sets the retrieval method
func (c *Client) SetRetrievalMethod(m RetrievalMethod) {
	c.retrievalMethod = m
}

// ReplicateToN sets how many nodes each key should be replicated to
func (c *Client) ReplicateToN(numNodes int) error {
	if numNodes > len(c.ch.Members()) {
		return errors.New("invalid number of nodes")
	}
	c.Lock()
	c.replicateNodeCt = numNodes
	c.Unlock()
	return nil
}

func (c *Client) node(nodeName string) cache.Cache {
	c.RLock()
	defer c.RUnlock()
	return c.nodes[nodeName]
}

// Set implements the "cache.Cache".Set() interface
func (c *Client) Set(key string, value interface{}, exp time.Duration) (err error) {

	nodes, err := c.ch.GetN(key, c.replicateNodeCt)
	if err != nil {
		return
	}

	if c.replicateMethod == ReplicateSync {
		var eg errgroup.Group
		for i := range nodes {
			nodeName := nodes[i]
			eg.Go(func() error {
				return c.node(nodeName).Set(key, value, exp)
			})
		}
		return eg.Wait()
	}

	err = c.node(nodes[0]).Set(key, value, exp)
	if len(nodes) > 1 {
		nodes = nodes[1:]
		for i := range nodes {
			go c.node(nodes[i]).Set(key, value, exp)
		}
	}

	return
}

// Get implements the "cache.Cache".Get() interface. It checks nodes in order
// of priority, and returns success if the value exists on any of them.
func (c *Client) Get(key string, dstVal interface{}) (err error) {
	nodes, err := c.ch.GetN(key, c.replicateNodeCt)
	if err != nil {
		return err
	}

	// asyncronously get the key from the first node which has it
	if c.retrievalMethod == RetrieveAsync {

		done := make(chan error, 1)
		timeout := make(chan struct{}, 1)

		c.RLock()
		to := c.timeoutGet
		c.RUnlock()
		if to > 0 {
			go func() {
				time.Sleep(to)
				timeout <- struct{}{}
			}()
		}

		go func() {
			var wg sync.WaitGroup
			wg.Add(len(c.nodes))
			for i := range c.nodes {
				go func(nm string) {
					// This is for unit testing, nobody would actually set a 1ns timeout, right?
					if to == time.Nanosecond {
						time.Sleep(time.Millisecond)
					}
					if e := c.node(nm).Get(key, dstVal); e == nil {
						done <- nil
					}
					wg.Done()
				}(i)
			}

			wg.Wait()
			done <- cache.ErrCacheMiss
		}()

		select {
		case <-timeout:
			return ErrGetTimeout
		case err = <-done:
			return err
		}
	}

	// syncronous mode
	for i := range nodes {
		if err = c.node(nodes[i]).Get(key, dstVal); err == nil {
			return
		}
	}

	return cache.ErrCacheMiss
}

// Exists implements the "cache.Cache".Exists() interface
func (c *Client) Exists(key string) (exists bool) {

	timeout := make(chan struct{}, 1)
	found := make(chan bool, 1)

	go func() {
		if c.timeoutGet > 0 {
			time.Sleep(c.timeoutGet)
			timeout <- struct{}{}
		}
	}()

	go func() {
		nodes, err := c.ch.GetN(key, c.replicateNodeCt)
		if err != nil {
			found <- false
			return
		}

		if c.retrievalMethod == RetrieveAsync {
			var wg sync.WaitGroup
			wg.Add(len(c.nodes))
			for i := range nodes {
				go func(nm string) {
					if c.node(nm).Exists(key) {
						found <- true
					}
					wg.Done()
				}(nodes[i])
			}
			wg.Wait()
			found <- false
			return
		}

		for i := range nodes {
			if c.node(nodes[i]).Exists(key) {
				found <- true
			}
		}
	}()

	select {
	case <-timeout:
		return false
	case exists = <-found:
		return
	}
}

// Del implements the "cache.Cache".Del() interface. It deletes the given key across
// all replicated nodes and returns error if any of those delete operations fail.
func (c *Client) Del(key string) (err error) {

	var nodes []string

	c.RLock()
	to := c.timeoutSet
	rn := c.replicateNodeCt
	nodes, err = c.ch.GetN(key, rn)
	if err != nil {
		return err
	}
	c.RUnlock()

	ec := make(chan error, 1)
	go func() {
		var eg errgroup.Group
		for i := range nodes {
			nm := nodes[i]
			eg.Go(func() error {
				if to == time.Nanosecond {
					time.Sleep(time.Millisecond)
				}
				return c.node(nm).Del(key)
			})
		}
		ec <- eg.Wait()
	}()

	select {
	case err = <-ec:
		return
	case <-time.After(to):
		return ErrSetTimeout
	}
}

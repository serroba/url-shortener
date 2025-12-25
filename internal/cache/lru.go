package cache

import (
	"sync"

	"github.com/serroba/web-demo-go/internal/shortener"
)

// node represents a doubly linked list node.
type node struct {
	key   string
	value *shortener.ShortURL
	prev  *node
	next  *node
}

// LRU implements a Least Recently Used cache.
// It uses a doubly linked list for ordering and a map for O(1) lookups.
type LRU struct {
	capacity int
	items    map[string]*node
	head     *node // sentinel - head.next is most recently used
	tail     *node // sentinel - tail.prev is least recently used
	mu       sync.RWMutex
}

// New creates a new LRU cache with the given capacity.
func New(capacity int) *LRU {
	head := &node{}
	tail := &node{}
	head.next = tail
	tail.prev = head

	return &LRU{
		capacity: capacity,
		items:    make(map[string]*node),
		head:     head,
		tail:     tail,
	}
}

// Get retrieves a value from the cache.
// Returns the value and true if found, nil and false otherwise.
// Accessing an item moves it to the front (most recently used).
func (c *LRU) Get(key string) (*shortener.ShortURL, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if n, ok := c.items[key]; ok {
		c.moveToFront(n)

		return n.value, true
	}

	return nil, false
}

// Set adds or updates a value in the cache.
// If the cache is at capacity, the least recently used item is evicted.
func (c *LRU) Set(key string, value *shortener.ShortURL) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if n, ok := c.items[key]; ok {
		n.value = value
		c.moveToFront(n)

		return
	}

	// Evict if at capacity
	if len(c.items) >= c.capacity {
		c.evictLRU()
	}

	// Add new node at front
	n := &node{key: key, value: value}
	c.items[key] = n
	c.addToFront(n)
}

// Len returns the current number of items in the cache.
func (c *LRU) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.items)
}

// moveToFront detaches a node and reattaches it at the front.
func (c *LRU) moveToFront(n *node) {
	c.detach(n)
	c.addToFront(n)
}

// addToFront adds a node right after the head sentinel.
func (c *LRU) addToFront(n *node) {
	n.prev = c.head
	n.next = c.head.next
	c.head.next.prev = n
	c.head.next = n
}

// detach removes a node from the list.
func (c *LRU) detach(n *node) {
	n.prev.next = n.next
	n.next.prev = n.prev
}

// evictLRU removes the least recently used item (right before tail sentinel).
func (c *LRU) evictLRU() {
	lru := c.tail.prev
	if lru == c.head {
		return // empty list
	}

	c.detach(lru)
	delete(c.items, lru.key)
}

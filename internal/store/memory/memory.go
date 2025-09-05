package memory

import (
	"container/list"
	"sync"
)

type entry struct {
	key   string
	value []byte
}

type Store struct {
	sync.RWMutex
	capacity int
	items    map[string]*list.Element
	order    *list.List // front = most recent, back = least recent
}

func New(capacity int) *Store {
	if capacity <= 0 {
		capacity = 10000
	}
	return &Store{
		capacity: capacity,
		items:    make(map[string]*list.Element, capacity),
		order:    list.New(),
	}
}

func (c *Store) Set(key string, value []byte) {
	c.Lock()
	defer c.Unlock()

	if elem, ok := c.items[key]; ok {
		// Update existing and move to front
		elem.Value.(*entry).value = value
		c.order.MoveToFront(elem)
		return
	}
	// Insert new at front
	e := &entry{key: key, value: value}
	elem := c.order.PushFront(e)
	c.items[key] = elem
	// Evict if over capacity
	if c.order.Len() > c.capacity {
		c.evict()
	}
}

func (c *Store) Get(key string) ([]byte, bool) {
	c.Lock()
	defer c.Unlock()

	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		return elem.Value.(*entry).value, true
	}
	return nil, false
}

func (c *Store) evict() {
	oldest := c.order.Back()
	if oldest == nil {
		return
	}
	c.order.Remove(oldest)
	kv := oldest.Value.(*entry)
	delete(c.items, kv.key)
}

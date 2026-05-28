package state

import "container/list"

// Cache keeps view-scoped state in least-recently-used order. It is intended
// for session memory only: scroll positions, toggles, focused panes, and other
// UI state that should survive navigation but not app restarts.
type Cache[V any] struct {
	limit int
	items map[string]*list.Element
	order *list.List
}

type entry[V any] struct {
	key   string
	value V
}

func NewCache[V any](limit int) *Cache[V] {
	if limit < 1 {
		limit = 1
	}
	return &Cache[V]{
		limit: limit,
		items: map[string]*list.Element{},
		order: list.New(),
	}
}

func (c *Cache[V]) Get(key string) (V, bool) {
	var zero V
	if c == nil || key == "" {
		return zero, false
	}
	el, ok := c.items[key]
	if !ok {
		return zero, false
	}
	c.order.MoveToFront(el)
	return el.Value.(*entry[V]).value, true
}

func (c *Cache[V]) Put(key string, value V) {
	if c == nil || key == "" {
		return
	}
	if el, ok := c.items[key]; ok {
		el.Value.(*entry[V]).value = value
		c.order.MoveToFront(el)
		return
	}
	el := c.order.PushFront(&entry[V]{key: key, value: value})
	c.items[key] = el
	c.evict()
}

func (c *Cache[V]) Delete(key string) {
	if c == nil || key == "" {
		return
	}
	el, ok := c.items[key]
	if !ok {
		return
	}
	c.order.Remove(el)
	delete(c.items, key)
}

func (c *Cache[V]) Len() int {
	if c == nil {
		return 0
	}
	return len(c.items)
}

func (c *Cache[V]) evict() {
	for len(c.items) > c.limit {
		el := c.order.Back()
		if el == nil {
			return
		}
		c.order.Remove(el)
		delete(c.items, el.Value.(*entry[V]).key)
	}
}

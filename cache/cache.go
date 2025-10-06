package cache

import (
	"sync"
	"time"
)

type CacheItem struct {
	Data      interface{}
	ExpiresAt time.Time
}

type Cache struct {
	items map[string]CacheItem
	mutex sync.RWMutex
}

var GlobalCache = &Cache{
	items: make(map[string]CacheItem),
}

// Set adiciona um item ao cache com duração
func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items[key] = CacheItem{
		Data:      value,
		ExpiresAt: time.Now().Add(duration),
	}
}

// Get recupera um item do cache se não estiver expirado
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.ExpiresAt) {
		// Item expirado, remover do cache
		go func() {
			c.mutex.Lock()
			delete(c.items, key)
			c.mutex.Unlock()
		}()
		return nil, false
	}

	return item.Data, true
}

// Clear remove todos os itens do cache
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.items = make(map[string]CacheItem)
}

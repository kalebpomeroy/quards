package services

import (
	"sync"
)

// InMemoryCache implements CacheService interface with in-memory storage
type InMemoryCache struct {
	data  map[string]interface{}
	mutex sync.RWMutex
	hits  int64
	misses int64
}

// NewInMemoryCache creates a new in-memory cache
func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{
		data: make(map[string]interface{}),
	}
}

// Get retrieves a value from the cache
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	value, exists := c.data[key]
	if exists {
		c.hits++
	} else {
		c.misses++
	}
	return value, exists
}

// Set stores a value in the cache
func (c *InMemoryCache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.data[key] = value
}

// Clear removes all items from the cache
func (c *InMemoryCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.data = make(map[string]interface{})
	c.hits = 0
	c.misses = 0
}

// GetStats returns cache statistics
func (c *InMemoryCache) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	total := c.hits + c.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(c.hits) / float64(total)
	}
	
	return map[string]interface{}{
		"hits":     c.hits,
		"misses":   c.misses,
		"total":    total,
		"hit_rate": hitRate,
		"size":     len(c.data),
	}
}
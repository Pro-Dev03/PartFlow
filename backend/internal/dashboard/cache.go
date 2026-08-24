package dashboard

import (
	"sync"
	"time"
)

// Cache represents a simple in-memory cache
type Cache struct {
	data      *DashboardStats
	timestamp time.Time
	mu        sync.RWMutex
	ttl       time.Duration
}

// NewCache creates a new cache with specified TTL
func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		ttl: ttl,
	}
}

// Get retrieves cached data if still valid
func (c *Cache) Get() (*DashboardStats, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.data == nil || time.Since(c.timestamp) > c.ttl {
		return nil, false
	}

	return c.data, true
}

// Set stores data in cache
func (c *Cache) Set(data *DashboardStats) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = data
	c.timestamp = time.Now()
}

// Clear clears the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = nil
	c.timestamp = time.Time{}
}

package dashboard

import (
	"sync"
	"time"
)

// EnhancedCache represents a cache for single-tenant system
type EnhancedCache struct {
	data      *CachedData // Single cache entry
	mu        sync.RWMutex
	ttl       time.Duration
}

type CachedData struct {
	stats     *DashboardStats
	timestamp time.Time
}

// NewEnhancedCache creates a new enhanced cache for single-tenant system
func NewEnhancedCache(ttl time.Duration) *EnhancedCache {
	return &EnhancedCache{
		data: nil,
		ttl:  ttl,
	}
}

// Get retrieves cached data if still valid
func (c *EnhancedCache) Get() (*DashboardStats, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cachedData := c.data
	if cachedData == nil {
		return nil, false
	}

	if time.Since(cachedData.timestamp) > c.ttl {
		return nil, false
	}

	return cachedData.stats, true
}

// Set stores data in cache
func (c *EnhancedCache) Set(data *DashboardStats) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = &CachedData{
		stats:     data,
		timestamp: time.Now(),
	}
}

// Clear clears cache
func (c *EnhancedCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = nil
}

// ClearAll clears all cache (use with caution)
func (c *EnhancedCache) ClearAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = nil
}

// Cleanup removes expired entries
func (c *EnhancedCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.data != nil && time.Since(c.data.timestamp) > c.ttl {
		c.data = nil
	}
}
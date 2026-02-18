package webhook

import (
	"sync"
	"time"
)

type verifierCache struct {
	mu      sync.RWMutex
	ttl     time.Duration
	entries map[string]time.Time
}

func newVerifierCache(ttl time.Duration) *verifierCache {
	if ttl <= 0 {
		return nil
	}
	return &verifierCache{
		ttl:     ttl,
		entries: make(map[string]time.Time),
	}
}

func (c *verifierCache) hasFresh(key string, now time.Time) bool {
	if c == nil {
		return false
	}
	c.mu.RLock()
	expiresAt, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return false
	}
	if expiresAt.After(now) {
		return true
	}
	c.mu.Lock()
	delete(c.entries, key)
	c.mu.Unlock()
	return false
}

func (c *verifierCache) putSuccess(key string, now time.Time) {
	if c == nil {
		return
	}
	c.mu.Lock()
	c.entries[key] = now.Add(c.ttl)
	c.mu.Unlock()
}

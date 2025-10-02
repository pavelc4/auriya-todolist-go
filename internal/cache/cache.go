package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type Service struct {
	client *cache.Cache
}

// NewService creates a new cache service with default expiration times.
func NewService(defaultExpiration, cleanupInterval time.Duration) *Service {
	return &Service{
		client: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Set adds an item to the cache, replacing any existing item.
func (s *Service) Set(key string, value interface{}, duration time.Duration) {
	s.client.Set(key, value, duration)
}

// Get retrieves an item from the cache.
func (s *Service) Get(key string) (interface{}, bool) {
	return s.client.Get(key)
}

// Delete removes an item from the cache.
func (s *Service) Delete(key string) {
	s.client.Delete(key)
}

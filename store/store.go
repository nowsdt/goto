package store

import (
	"goto/arith"
	"sync"
)

type URLStore struct {
	urls map[string]string
	mu sync.RWMutex
}

func (s *URLStore) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.urls[key]
}

func (s *URLStore) Set(key, url string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, present := s.urls[key]; present {
		return false
	}

	s.urls[key] = url
	return true
}

func (s *URLStore) Put(url string) string  {
	for {
		key := arith.Short(len(url))
		if s.Set(key, url) {
			return key
		}
	}

	return ""
}

func (s *URLStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.urls)
}

func NewURLStore() *URLStore  {
	return &URLStore{urls: make(map[string]string)}
}
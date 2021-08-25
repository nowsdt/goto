package store

import (
	"encoding/json"
	"fmt"
	"goto/arith"
	"goto/utils"
	"io"
	"log"
	"os"
	"sync"
)

const saveQueueLength = 1000

type URLStore struct {
	urls map[string]string
	mu   sync.RWMutex
	save chan record
}

type record struct {
	Key, URL string
}

func (s *URLStore) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.urls[key]
}

func (s *URLStore) Set(key, url string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if utils.ContainsValue(s.urls, url) {
		return false
	}

	s.urls[key] = url
	return true
}

func (s *URLStore) Put(url string) string {
	for {
		key := arith.Short(s.Count())
		if s.Set(key, url) {
			s.save <- record{key, url}
			return key
		} else {
			return fmt.Sprintf("%s exists", url)
		}
	}

	panic("shouldn’t get here")
}

func (s *URLStore) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.urls)
}

func NewURLStore(filename string) *URLStore {
	s := &URLStore{
		urls: make(map[string]string),
		save: make(chan record, saveQueueLength),
	}
	if err := s.load(filename); err != nil {
		log.Println("Error loading data in URLStore:", err)
	}
	go s.saveLoop(filename)

	return s
}

func (s *URLStore) saveLoop(filename string) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal("URLStore:", err)
	}
	defer f.Close()
	e := json.NewEncoder(f)

	for {
		r := <-s.save
		if err := e.Encode(r); err != nil {
			log.Println("URLStore:", err)
		}
	}
}

func (s *URLStore) load(filename string) error {
	var err error

	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(f)

	for err == nil {
		var r record
		if err = decoder.Decode(&r); err == nil {
			s.Set(r.Key, r.URL)
		}
	}
	if err == io.EOF {
		return nil
	}

	return err
}

package store

import (
	"encoding/json"
	"errors"
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

type Store interface {
	Put(url, key *string) error
	Get(key, url *string) error
}

type record struct {
	Key, URL string
}

func (s *URLStore) Get(key, url *string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if u, ok := s.urls[*key]; ok {
		*url = u
		return nil
	}
	return errors.New("key not found")
}

func (s *URLStore) Set(key, url *string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if utils.ContainsValue(s.urls, *url) {
		return errors.New("key already exists")
	}

	s.urls[*key] = *url
	return nil
}

func (s *URLStore) Put(url, key *string) error {
	var err error
	for {
		*key = arith.Short(s.count())
		if err = s.Set(key, url); err == nil {
			break
		} else {
			return err
		}
	}

	if s.save != nil {
		s.save <- record{*key, *url}
	}
	return nil
}

func (s *URLStore) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.urls)
}

func NewURLStore(filename string) *URLStore {
	s := &URLStore{
		urls: make(map[string]string),
	}
	if len(filename) != 0 {
		s.save = make(chan record, saveQueueLength)
		if err := s.load(filename); err != nil {
			log.Println("Error loading data in URLStore:", err)
		}
		go s.saveLoop(filename)
	}

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

	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(f)

	for err == nil {
		var r record
		if err = decoder.Decode(&r); err == nil {
			s.Set(&r.Key, &r.URL)
			log.Printf("k:%s,v:%s loaded\n", r.Key, r.URL)
		}
	}
	if err == io.EOF {
		return nil
	}

	return err
}

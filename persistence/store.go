package persistence

import (
	"sync"
	"time"
)

type value struct {
	expiration *time.Time
	val        []byte
}

type Store struct {
	data map[string]value
	mu   sync.Mutex
}

func NewStore() *Store {
	return &Store{
		data: map[string]value{},
		mu:   sync.Mutex{},
	}
}

func (s *Store) Set(key string, val []byte, expiration *time.Time) {
	defer s.mu.Unlock()

	s.mu.Lock()
	s.data[key] = value{
		expiration: expiration,
		val:        val,
	}
}

func (s *Store) Get(key string) ([]byte, bool) {
	defer s.mu.Unlock()

	s.mu.Lock()
	if val, ok := s.data[key]; ok {
		// Check if the value is expired
		if val.expiration != nil && time.Now().After(*val.expiration) {
			// Delete the value if its expired
			delete(s.data, key)

			return []byte{}, false
		}

		return val.val, true
	}

	// No value found
	return []byte{}, false
}

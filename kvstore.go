package main

import (
	//"encoding/gob"
	"log"
	//"strings"
	"sync"
)

// a key-value store backed by raft
type KVStore struct {
	mu		 sync.RWMutex
	data     map[string]string // current committed key-value pairs
	proposeC	 chan<- KV
}

type KV struct {
	Key string
	Val string
}

func NewKVStore(proposeC chan<- KV) *KVStore {
	return &KVStore{ data: make(map[string]string), proposeC: proposeC }
}

func (s *KVStore) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	v, ok := s.data[key]
	return v, ok
}

func (s *KVStore) ProposeSet(key string, val string) {
	//var buf strings.Builder

	//if err := gob.NewEncoder(&buf).Encode(KV{key, val}); err != nil {
	//	log.Fatal(err)
	//}

	log.Println("going to set")
	//s.proposeC <- buf.String()
	s.proposeC <- KV{key, val}
	log.Println("set")
}

func (s *KVStore) Set(k string, v string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[k] = v
}

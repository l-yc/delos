package main

import (
	"context"
	//"encoding/gob"
	"log"
	//"strings"
	"sync"
)

// a key-value store backed by raft
type KVStore struct {
	mu		 sync.RWMutex
	data     map[string]string // current committed key-value pairs
	engine	 *IEngine[string, Entry]
}

type KV struct {
	Key string
	Val string
}

func NewKVStore(engine *IEngine[string, Entry]) KVStore {
	kvs := KVStore{ data: make(map[string]string), engine: engine }
	log.Println("created kv store", kvs)

	var test IApplicator[string, Entry] = kvs
	(*engine).RegisterUpcall(&test)
	log.Println("registered upcall for engine")
	return kvs
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

	(*s.engine).Propose(context.TODO(), Entry{ Data: KV{key, val} })
}

func (s *KVStore) Set(k string, v string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[k] = v
}


// wrapper?
func (s KVStore) Apply(txn RWTx, e Entry, pos LogPos) string {
	log.Println("apply kv", e)
	entry := e
	data := entry.Data.(KV)
	s.Set(data.Key, data.Val)
	return "ok"
}

func (s KVStore) PostApply(e Entry, pos LogPos) {

}

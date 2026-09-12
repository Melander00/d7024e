package kademlia

import (
	"maps"
	"slices"
	"sync"
)

type DataStore struct {
	store map[*KademliaID][]byte
	mu    sync.RWMutex
}

func NewDataStore() *DataStore {
	return &DataStore{
		store: make(map[*KademliaID][]byte),
	}
}

func (ds *DataStore) Put(key *KademliaID, value []byte) error {
	ds.mu.Lock()
	ds.store[key] = value
	ds.mu.Unlock()
	return nil
}

func (ds *DataStore) Get(key *KademliaID) ([]byte, bool) {
	ds.mu.RLock()
	data, exists := ds.store[key]
	ds.mu.RUnlock()
	return data, exists
}

func (ds *DataStore) Keys() []*KademliaID {
	ds.mu.RLock()
	data := slices.Collect(maps.Keys(ds.store))
	ds.mu.RUnlock()
	return data
}

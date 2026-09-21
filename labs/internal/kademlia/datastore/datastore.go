package datastore

import (
	"d7024e/internal/kademlia/contact"
	"sync"
)

type DataStore struct {
	store map[contact.KademliaID][]byte
	mu    sync.RWMutex
}

func NewDataStore() *DataStore {
	return &DataStore{
		store: make(map[contact.KademliaID][]byte),
	}
}

func (ds *DataStore) Put(key *contact.KademliaID, value []byte) error {
	ds.mu.Lock()
	ds.store[*key] = value
	ds.mu.Unlock()
	return nil
}

func (ds *DataStore) Get(key *contact.KademliaID) ([]byte, bool) {
	ds.mu.RLock()
	data, exists := ds.store[*key]
	ds.mu.RUnlock()
	return data, exists
}

func (ds *DataStore) Keys() []*contact.KademliaID {
	ds.mu.RLock()
	data := make([]*contact.KademliaID, 0, len(ds.store))
	for key := range ds.store {
		keyCopy := key
		data = append(data, &keyCopy)
	}
	ds.mu.RUnlock()
	return data
}

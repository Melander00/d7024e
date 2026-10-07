package kademlia

import (
	"context"
	"d7024e/internal/kademlia/contact"
	"time"
)

func (kademlia *Kademlia) replicateValue(key *contact.KademliaID, data []byte) {
	target := contact.NewContact(key, "")
	contacts := kademlia.LookupContact(&target)

	for _, candidate := range contacts {
		if candidate.Address == kademlia.Me.Address {
			continue
		}
		_, err := kademlia.Rpc.Store(candidate, key.String(), data)
		if err != nil {
			continue // try other contacts; future ticks retry replication
		}
	}
}

func (kademlia *Kademlia) replicateOnce() {
	for _, key := range kademlia.Datastore.Keys() {
		value, exists := kademlia.Datastore.Get(key)
		if !exists {
			continue // skip keys that no longer exist
		}
		expectedKey := contact.NewKademliaIDFromData(value)
		if !key.Equals(expectedKey) {
			continue // skip keys that no longer match their expected hash
		}
		kademlia.replicateValue(key, value)
	}
}

func (kademlia *Kademlia) runReplicationLoop(ctx context.Context) {
	ticker := time.NewTicker(kademlia.Config.ReplicationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			kademlia.replicateOnce()
		}
	}
}

func (kademlia *Kademlia) StartReplication(ctx context.Context) {
	go kademlia.runReplicationLoop(ctx)
}

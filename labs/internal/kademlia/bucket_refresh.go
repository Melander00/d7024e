package kademlia

import (
	"context"
	"crypto/rand"
	"d7024e/internal/kademlia/contact"
	"time"
)

// randomIDForBucket creates a random Kademlia ID whose XOR distance from
// our own ID has its first set bit at the requested bucket index.
func randomIDForBucket(me *contact.KademliaID, bucketIndex int) (*contact.KademliaID, error) {

	distance := contact.KademliaID{}
	byteIndex := bucketIndex / 8
	bitIndex := bucketIndex % 8

	randomBytes := make([]byte, contact.IDLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, err
	}

	for i := 0; i < byteIndex; i++ {
		distance[i] = 0
	}

	distance[byteIndex] = 1 << (7 - bitIndex)
	for bit := bitIndex + 1; bit < 8; bit++ {
		if randomBytes[byteIndex]&(1<<(7-bit)) != 0 {
			distance[byteIndex] |= 1 << (7 - bit)
		}
	}

	for i := byteIndex + 1; i < contact.IDLength; i++ {
		distance[i] = randomBytes[i]
	}

	target := *me
	for i := 0; i < contact.IDLength; i++ {
		target[i] ^= distance[i]
	}

	return &target, nil
}

// refreshBucket performs a node lookup for a random ID in the
// requested bucket's XOR-distance range.
func (kademlia *Kademlia) refreshBucket(bucketIndex int) {

	targetID, err := randomIDForBucket(
		kademlia.ID,
		bucketIndex)
	// fix ts eventually
	if err != nil {
		return
	}

	target := contact.NewContact(targetID, "")

	kademlia.LookupContact(&target)

}

// Currently the stale threshold and scan frequency are the same setting,
// which means refresh isn't exact to the threshold.
// chosen initially for simplicity

func (kademlia *Kademlia) refreshStaleBuckets() {

	staleBuckets :=
		kademlia.Routing.BucketsNeedingRefresh(
			kademlia.Config.BucketRefreshInterval,
		)

	for _, bucketIndex := range staleBuckets {
		kademlia.refreshBucket(bucketIndex)
	}

}

func (kademlia *Kademlia) runBucketRefreshLoop(ctx context.Context) {
	ticker := time.NewTicker(
		kademlia.Config.BucketRefreshInterval,
	)
	defer ticker.Stop()

	for {
		select {

		case <-ticker.C:
			kademlia.refreshStaleBuckets()

		case <-ctx.Done():
			return
		}
	}

}

// StartBucketRefresh starts the background loop that periodically refreshes stale buckets.
func (kademlia *Kademlia) StartBucketRefresh(ctx context.Context) {
	go kademlia.runBucketRefreshLoop(ctx)
}

package contact

import (
	"sync"
	"time"
)

// RoutingTable definition
// keeps a refrence contact of me and an array of buckets
type RoutingTable struct {
	me         Contact
	buckets    [IDLength * 8]*bucket
	bucketSize int

	mu sync.RWMutex
}

// NewRoutingTable returns a new instance of a RoutingTable
func NewRoutingTable(me Contact, bucketSize int) *RoutingTable {
	routingTable := &RoutingTable{
		bucketSize: bucketSize,
	}
	for i := 0; i < IDLength*8; i++ {
		routingTable.buckets[i] = newBucket(bucketSize)
	}
	routingTable.me = me
	return routingTable
}

// AddContact add a new contact to the correct Bucket
func (routingTable *RoutingTable) AddContact(contact Contact) {
	routingTable.mu.Lock()
	defer routingTable.mu.Unlock()
	if contact.ID.Equals(routingTable.me.ID) {
		return
	}
	bucketIndex := routingTable.getBucketIndex(contact.ID)
	bucket := routingTable.buckets[bucketIndex]
	bucket.AddContact(contact)
}

// Remove a contact from the corresponding Bucket.
func (routingTable *RoutingTable) RemoveContact(contact Contact) {
	routingTable.mu.Lock()
	defer routingTable.mu.Unlock()

	bucketIndex := routingTable.getBucketIndex(contact.ID)
	bucket := routingTable.buckets[bucketIndex]
	bucket.RemoveContact(contact)
}

// FindClosestContacts finds the count closest Contacts to the target in the RoutingTable
func (routingTable *RoutingTable) FindClosestContacts(target *KademliaID, count int) []Contact {
	routingTable.mu.RLock()
	defer routingTable.mu.RUnlock()

	var candidates ContactCandidates
	bucketIndex := routingTable.getBucketIndex(target)
	bucket := routingTable.buckets[bucketIndex]

	candidates.Append(bucket.GetContactAndCalcDistance(target))

	for i := 1; (bucketIndex-i >= 0 || bucketIndex+i < IDLength*8) && candidates.Len() < count; i++ {
		if bucketIndex-i >= 0 {
			bucket = routingTable.buckets[bucketIndex-i]
			candidates.Append(bucket.GetContactAndCalcDistance(target))
		}
		if bucketIndex+i < IDLength*8 {
			bucket = routingTable.buckets[bucketIndex+i]
			candidates.Append(bucket.GetContactAndCalcDistance(target))
		}
	}

	candidates.Sort()

	if count > candidates.Len() {
		count = candidates.Len()
	}

	return candidates.GetContacts(count)
}

// getBucketIndex get the correct Bucket index for the KademliaID
func (routingTable *RoutingTable) getBucketIndex(id *KademliaID) int {
	distance := id.CalcDistance(routingTable.me.ID)
	for i := 0; i < IDLength; i++ {
		for j := 0; j < 8; j++ {
			if (distance[i]>>uint8(7-j))&0x1 != 0 {
				return i*8 + j
			}
		}
	}

	return IDLength*8 - 1
}

func (routingTable *RoutingTable) BucketIndex(
	id *KademliaID,
) int {
	return routingTable.getBucketIndex(id)
}

func (routingTable *RoutingTable) GetBuckets() [256]*bucket {
	return routingTable.buckets
}

// MarkLookup marks the last lookup time for the bucket corresponding to the target KademliaID
func (routingTable *RoutingTable) MarkLookup(target *KademliaID) {
	bucketIndex := routingTable.getBucketIndex(target)

	bucket := routingTable.buckets[bucketIndex]

	bucket.MarkLookup()
}

// BucketsNeedingRefresh returns a list of bucket indices that need to be refreshed based on the given interval.
func (routingTable *RoutingTable) BucketsNeedingRefresh(
	interval time.Duration,
) []int {

	stale := []int{}

	for index, bucket := range routingTable.buckets {
		if bucket.NeedsRefresh(interval) {
			stale = append(stale, index)
		}
	}

	return stale
}

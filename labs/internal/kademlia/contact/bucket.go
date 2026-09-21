package contact

import (
	"container/list"
	"sync"
	"time"
)

// bucket definition
// contains a List
type bucket struct {
	list       *list.List
	size       int
	lastLookup time.Time
	mu         sync.RWMutex
}

// newBucket returns a new instance of a bucket
func newBucket(size int) *bucket {
	bucket := &bucket{
		size: size,
	}
	bucket.list = list.New()
	return bucket
}

// AddContact adds the Contact to the front of the bucket
// or moves it to the front of the bucket if it already existed
func (bucket *bucket) AddContact(contact Contact) {
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	var element *list.Element
	for e := bucket.list.Front(); e != nil; e = e.Next() {
		nodeID := e.Value.(Contact).ID

		if (contact).ID.Equals(nodeID) {
			element = e
		}
	}

	if element == nil {
		if bucket.list.Len() < bucket.size {
			bucket.list.PushFront(contact)
		}
	} else {
		bucket.list.MoveToFront(element)
	}
}

// Remove contact from this bucket.
func (bucket *bucket) RemoveContact(contact Contact) {
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	var element *list.Element
	for e := bucket.list.Front(); e != nil; e = e.Next() {
		nodeID := e.Value.(Contact).ID

		if (contact).ID.Equals(nodeID) {
			element = e
		}
	}

	if element != nil {
		bucket.list.Remove(element)
	}
}

// GetContactAndCalcDistance returns an array of Contacts where
// the distance has already been calculated
func (bucket *bucket) GetContactAndCalcDistance(target *KademliaID) []Contact {
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	var contacts []Contact

	for elt := bucket.list.Front(); elt != nil; elt = elt.Next() {
		contact := elt.Value.(Contact)
		contact.CalcDistance(target)
		contacts = append(contacts, contact)
	}

	return contacts
}

// Len return the size of the bucket
func (bucket *bucket) Len() int {
	return bucket.list.Len()
}

func (bucket *bucket) GetContacts() []Contact {
	bucket.mu.RLock()
	defer bucket.mu.RUnlock()

	var contacts []Contact

	for elt := bucket.list.Front(); elt != nil; elt = elt.Next() {
		contact := elt.Value.(Contact)
		contacts = append(contacts, contact)
	}

	return contacts
}

// MarkLookup updates the last lookup time for the bucket to the current time.
func (bucket *bucket) MarkLookup() {
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	bucket.lastLookup = time.Now()
}

// NeedsRefresh checks if the bucket needs to be refreshed based on the given interval.
func (bucket *bucket) NeedsRefresh(interval time.Duration) bool {
	bucket.mu.RLock()
	defer bucket.mu.RUnlock()

	if bucket.lastLookup.IsZero() {
		return true
	}

	return time.Since(bucket.lastLookup) >= interval
}

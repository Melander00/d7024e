package kademlia

import (
	"d7024e/internal/kademlia/contact"
	"d7024e/internal/kademlia/datastore"
	"d7024e/internal/kademlia/rpc"
	"time"
)

type Kademlia struct {
	Rpc       *rpc.RPC
	ID        *contact.KademliaID
	Routing   *contact.RoutingTable
	Me        contact.Contact
	Datastore *datastore.DataStore
	Config    *KademliaConfig
}

type KademliaConfig struct {
	Alpha                 int
	K                     int
	Timeout               time.Duration
	Retries               int
	BucketRefreshInterval time.Duration
}

// defaultBucketRefreshInterval is used when a config leaves BucketRefreshInterval unset. set to half an hour to guarantee periodic refresh of buckets of an hour
const defaultBucketRefreshInterval = time.Hour

func NewKademliaNode(config KademliaConfig, rpc *rpc.RPC) *Kademlia {

	id, _ := contact.NewKademliaIDFromAddress(rpc.Me.Address)

	rpc.SetTimeout(config.Timeout)
	rpc.SetRetries(config.Retries)

	if config.BucketRefreshInterval <= 0 {
		config.BucketRefreshInterval = defaultBucketRefreshInterval
	}

	routing := contact.NewRoutingTable(rpc.Me, config.K)

	kademlia := &Kademlia{
		Rpc:       rpc,
		ID:        id,
		Me:        rpc.Me,
		Routing:   routing,
		Datastore: datastore.NewDataStore(),
		Config:    &config,
	}

	rpc.SetPingHandler(kademlia.pingHandler)
	rpc.SetFindNodeHandler(kademlia.findNodeHandler)
	rpc.SetFindValueHandler(kademlia.findValueHandler)
	rpc.SetStoreHandler(kademlia.storeHandler)

	return kademlia
}

type lookupResult struct {
	contacts []contact.Contact
	value    []byte
	found    bool
	err      error
	contact  contact.Contact
}

type lookupQuery func(contact.Contact, *contact.KademliaID) lookupResult

func (kademlia *Kademlia) lookup(target *contact.KademliaID, query lookupQuery) ([]contact.Contact, []byte, error) {
	candidates := contact.ContactCandidates{}
	candidates.Append(kademlia.Routing.FindClosestContacts(target, kademlia.Config.K))

	candidates.RemoveMe(&kademlia.Me)

	candidates.Sort()

	queried := make(map[string]bool)
	results := make(chan lookupResult, kademlia.Config.Alpha)

	pending := 0

	for {

		for pending < kademlia.Config.Alpha {

			candidate, ok := getNextCandidate(&candidates, queried, kademlia.Config.K)

			if !ok {
				// We have reached the end
				break
			}

			queried[candidate.ID.String()] = true
			pending++

			go func(candidate contact.Contact) {
				results <- query(candidate, target)
			}(candidate)
		}

		if pending == 0 {
			// If we have no more contacts to request and no more inflight requests we stop
			break
		}

		result := <-results
		pending--

		if result.err != nil {
			// TODO: error handling
			// If the error is timeout we can assume it is dead for example
			kademlia.Routing.RemoveContact(result.contact)
			continue
		}

		for _, c := range result.contacts {
			kademlia.Routing.AddContact(c)
		}
		kademlia.Routing.AddContact(result.contact)

		if result.found {
			return nil, result.value, nil
		}

		candidates.Append(result.contacts)
		candidates.SortNear(target)
	}

	// Return the k closest known contacts.
	if candidates.Len() > kademlia.Config.K {
		return candidates.GetContacts(kademlia.Config.K), nil, nil
	}

	return candidates.GetContacts(candidates.Len()), nil, nil
}

func getNextCandidate(candidates *contact.ContactCandidates, queried map[string]bool, k int) (contact.Contact, bool) {
	contacts := candidates.GetContacts(candidates.Len())

	for i, candidate := range contacts {
		if !queried[candidate.ID.String()] {
			return candidate, true
		}
		if i >= k {
			break
		}
	}

	return contact.Contact{}, false
}

func (kademlia *Kademlia) LookupContact(target *contact.Contact) []contact.Contact {
	// Self-lookups have no bucket (XOR distance is zero)
	if !target.ID.Equals(kademlia.ID) {
		kademlia.Routing.MarkLookup(target.ID)
	}

	contacts, _, _ := kademlia.lookup(
		target.ID,
		func(candidate contact.Contact, target *contact.KademliaID) lookupResult {
			res, err := kademlia.Rpc.FindNode(candidate, target.String())

			if err != nil {
				return lookupResult{
					contact: candidate,
					err:     err,
				}
			}

			return lookupResult{
				contact:  candidate,
				contacts: res.Value.Nodes,
			}
		},
	)

	return contacts
}

func (kademlia *Kademlia) LookupData(hash string) ([]byte, error) {
	target, err := contact.ParseKademliaID(hash)

	if err != nil {
		return nil, err
	}

	_, value, err := kademlia.lookup(
		target,
		func(candidate contact.Contact, target *contact.KademliaID) lookupResult {
			res, err := kademlia.Rpc.FindValue(candidate, target.String())

			if err != nil {
				return lookupResult{
					contact: candidate,
					err:     err,
				}
			}

			if res.Value.Value != nil {
				return lookupResult{
					contact: candidate,
					value:   res.Value.Value,
					found:   true,
				}
			}

			return lookupResult{
				contact:  candidate,
				contacts: res.Value.Nodes,
			}
		},
	)

	return value, err
}

func (kademlia *Kademlia) Store(data []byte) {
	// TODO
}

func (kademlia *Kademlia) pingHandler(client contact.Contact) {
	kademlia.Routing.AddContact(client)
	return // Ping shouldnt do anything.
}

func (kademlia *Kademlia) findNodeHandler(
	client contact.Contact,
	hash string,
) ([]contact.Contact, error) {

	target, err := contact.ParseKademliaID(hash)

	if err != nil {
		return nil, err
	}

	kademlia.Routing.AddContact(client)

	contacts :=
		kademlia.Routing.FindClosestContacts(
			target,
			kademlia.Config.K,
		)

	return contacts, nil
}

func (kademlia *Kademlia) findValueHandler(
	client contact.Contact,
	hash string,
) ([]contact.Contact, []byte, bool, error) {

	key, err := contact.ParseKademliaID(hash)

	if err != nil {
		return nil, nil, false, err
	}

	kademlia.Routing.AddContact(client)

	data, has := kademlia.Datastore.Get(key)

	if has {
		return nil, data, true, nil
	}

	contacts :=
		kademlia.Routing.FindClosestContacts(
			key,
			kademlia.Config.K,
		)

	return contacts, nil, false, nil
}

func (kademlia *Kademlia) storeHandler(client contact.Contact, key string, value []byte) bool {
	kademlia.Routing.AddContact(client)
	target, err := contact.ParseKademliaID(key)
	if err != nil {
		return false
	}

	expected := contact.NewKademliaIDFromData(value)
	if !target.Equals(expected) {
		return false
	}

	return kademlia.Datastore.Put(target, value) == nil
}

func (kademlia *Kademlia) Join(boot contact.Contact) {
	kademlia.Routing.AddContact(boot)
	kademlia.LookupContact(&kademlia.Me)
	// Find closest node we now know.
	closest := kademlia.Routing.FindClosestContacts(
		kademlia.ID,
		1,
	)
	// TODO: return error
	if len(closest) == 0 {
		return
	}

	// Find the bucket containing that closest neighbor.
	closestBucket :=
		kademlia.Routing.BucketIndex(closest[0].ID)

	// Our indexing has farther buckets at smaller indexes.
	for bucketIndex := 0; bucketIndex < closestBucket; bucketIndex++ {

		kademlia.refreshBucket(bucketIndex)
	}

}

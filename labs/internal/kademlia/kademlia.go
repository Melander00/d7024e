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
	Alpha   int
	K       int
	Timeout time.Duration
	Retries int
}

func NewKademliaNode(config KademliaConfig, rpc *rpc.RPC) *Kademlia {

	id, _ := contact.NewKademliaIDFromAddress(rpc.Me.Address)

	rpc.SetTimeout(config.Timeout)
	rpc.SetRetries(config.Retries)

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

			candidate, ok := getNextCandidate(&candidates, queried)

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

func getNextCandidate(candidates *contact.ContactCandidates, queried map[string]bool) (contact.Contact, bool) {
	contacts := candidates.GetContacts(candidates.Len())

	for _, candidate := range contacts {
		if !queried[candidate.ID.String()] {
			return candidate, true
		}
	}

	return contact.Contact{}, false
}

func (kademlia *Kademlia) LookupContact(target *contact.Contact) []contact.Contact {
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

func (kademlia *Kademlia) LookupData(hash string) []byte {
	target, _ := contact.ParseKademliaID(hash)

	_, value, _ := kademlia.lookup(
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

	return value
}

func (kademlia *Kademlia) Store(data []byte) {
	// TODO
}

func (kademlia *Kademlia) pingHandler(client contact.Contact) {
	kademlia.Routing.AddContact(client)
	return // Ping shouldnt do anything.
}

func (kademlia *Kademlia) findNodeHandler(client contact.Contact, hash string) []contact.Contact {
	kademlia.Routing.AddContact(client)
	target, _ := contact.ParseKademliaID(hash) // TODO: error handling

	return kademlia.Routing.FindClosestContacts(target, kademlia.Config.K)
}

func (kademlia *Kademlia) findValueHandler(client contact.Contact, hash string) ([]contact.Contact, []byte, bool) {
	kademlia.Routing.AddContact(client)
	key, _ := contact.ParseKademliaID(hash) // TODO: error handling

	data, has := kademlia.Datastore.Get(key)
	if has {
		return nil, data, true
	}

	return kademlia.Routing.FindClosestContacts(key, kademlia.Config.K), nil, false
}

func (kademlia *Kademlia) storeHandler(client contact.Contact, key string, value []byte) bool {
	kademlia.Routing.AddContact(client)
	// TODO: Implement
	return false
}

func (kademlia *Kademlia) Join(boot contact.Contact) {
	kademlia.Routing.AddContact(boot)
	kademlia.LookupContact(&kademlia.Me)
}

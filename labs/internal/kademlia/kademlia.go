package kademlia

import (
	"d7024e/internal/kademlia/rpc"
	"time"
)

type Kademlia struct {
	Rpc       *rpc.RPC
	ID        *KademliaID
	Routing   *RoutingTable
	Me        Contact
	Datastore *DataStore
}

type KademliaConfig struct {
	alpha   int
	k       int
	timeout time.Duration
	retries int
}

func NewKademliaNode(config KademliaConfig, rpc *rpc.RPC) *Kademlia {

	id, _ := NewKademliaIDFromAddress(string(rpc.Me))

	rpc.SetTimeout(config.timeout)
	rpc.SetRetries(config.retries)

	me := NewContact(id, string(rpc.Me))

	routing := NewRoutingTable(me, config.k)

	return &Kademlia{
		Rpc:       rpc,
		ID:        id,
		Me:        me,
		Routing:   routing,
		Datastore: NewDataStore(),
	}
}

func (kademlia *Kademlia) LookupContact(target *Contact) {
	// TODO
}

func (kademlia *Kademlia) LookupData(hash string) {
	// TODO
}

func (kademlia *Kademlia) Store(data []byte) {
	// TODO
}

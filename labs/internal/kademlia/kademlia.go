package kademlia

import "d7024e/internal/kademlia/rpc"

type Kademlia struct {
	rpc *rpc.RPC
}

func NewKademliaNode(rpc *rpc.RPC) *Kademlia {
	return &Kademlia{
		rpc: rpc,
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

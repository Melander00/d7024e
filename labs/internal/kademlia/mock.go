package kademlia

import (
	"d7024e/internal/kademlia/rpc"
	"d7024e/pkg/network"
	"fmt"
)

func MockKademlia(nrNodes int) {
	fmt.Printf("Mocking kademlia %d nodes", nrNodes)

	simulation := network.NewSimulation()

	for i := 0; i < nrNodes; i++ {

		// address :=

		receiver := network.NewChannelNetworkReceiver(5)
		net := simulation.NewSimulatedNetwork(receiver)
		net.Listen("127.0.0.1", 10000+i)
		rpc := rpc.CreateRpc(receiver, net, net.Address)
		node := NewKademliaNode(rpc)

		node.rpc.FindNode(net.Address, "ADBCDED")

	}

	// Block
	for {

	}
}

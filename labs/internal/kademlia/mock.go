package kademlia

import (
	"d7024e/internal/kademlia/rpc"
	"d7024e/pkg/network"
	"fmt"
	"strconv"
)

func MockKademlia(nrNodes int) {
	fmt.Printf("Mocking kademlia %d nodes", nrNodes)

	simulation := network.NewSimulation()

	for i := 0; i < nrNodes; i++ {

		// address :=

		receiver := network.NewChannelNetworkReceiver(5)
		net := simulation.NewSimulatedNetwork(receiver)
		net.Listen(network.Address("node-" + strconv.Itoa(i)))
		rpc := rpc.CreateRpc(receiver, net, net.Address)
		node := NewKademliaNode(rpc)

		node.rpc.FindNode(net.Address, "ADBCDED")

	}

	// Block
	for {

	}
}

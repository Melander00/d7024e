package kademlia

import (
	"d7024e/internal/kademlia/rpc"
	"d7024e/pkg/network"
	"fmt"
	"strconv"
	"time"
)

func MockKademlia(nrNodes int) {
	fmt.Printf("Mocking kademlia %d nodes\n", nrNodes)

	simulation := network.NewSimulation()

	config := KademliaConfig{
		alpha:   1,
		k:       10,
		timeout: 5 * time.Second,
		retries: 5,
	}

	// Bootstrap node
	boot_receiver := network.NewChannelNetworkReceiver(5)
	boot_net := simulation.NewSimulatedNetwork(boot_receiver)
	boot_net.Listen(network.Address("127.0.0.1" + ":" + strconv.Itoa(1000)))
	boot_rpc := rpc.CreateRpc(boot_receiver, boot_net, boot_net.Address)
	// NewKademliaNode(config, boot_rpc)
	boot_node := NewKademliaNode(config, boot_rpc)

	for i := 1; i < nrNodes; i++ {

		receiver := network.NewChannelNetworkReceiver(5)
		net := simulation.NewSimulatedNetwork(receiver)
		net.Listen(network.Address("127.0.0.1" + ":" + strconv.Itoa(1000+i)))
		rpc := rpc.CreateRpc(receiver, net, net.Address)
		node := NewKademliaNode(config, rpc)
		node.Routing.AddContact(boot_node.Me)

		go func() {
			res, _ := rpc.Ping(boot_net.Address)
			fmt.Printf("[%s] PONG %s\n", net.Address, res.RequestID)
		}()
	}

	// Block
	for {

	}
}

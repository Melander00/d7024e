package kademlia

import (
	"d7024e/internal/kademlia/contact"
	"d7024e/internal/kademlia/rpc"
	"d7024e/pkg/network"
	"fmt"
	"strconv"
	"time"
)

type Mock struct {
	simulation *network.Simulation
	config     KademliaConfig
	nodes      []*Kademlia
}

func MockKademlia(nrNodes int) {
	fmt.Printf("Mocking kademlia %d nodes\n", nrNodes)

	simulation := network.NewSimulation()

	config := KademliaConfig{
		alpha:   1,
		k:       10,
		timeout: 5 * time.Second,
		retries: 5,
	}

	mock := &Mock{
		simulation: simulation,
		config:     config,
	}

	boot_node := mock.createNode(0, simulation, config)

	for i := 1; i < nrNodes; i++ {
		node := mock.createNode(i, simulation, config)
		go node.Join(boot_node.Me)
	}

	time.Sleep(1 * time.Second)

	nodes := boot_node.Routing.FindClosestContacts(boot_node.Me.ID, nrNodes)
	fmt.Printf("Boot has %d nodes in routing table\n", len(nodes))
	nodes = mock.nodes[1].Routing.FindClosestContacts(boot_node.Me.ID, nrNodes)
	fmt.Printf("Node-10001 has %d nodes\n", len(nodes))
	for _, node := range nodes {

		fmt.Printf("\t%s %s \n", node.Address, node.ID)
	}
}

func (mock *Mock) createNode(i int, simulation *network.Simulation, config KademliaConfig) *Kademlia {
	receiver := network.NewChannelNetworkReceiver(5)
	network := simulation.NewSimulatedNetwork(receiver)
	address := "127.0.0.1" + ":" + strconv.Itoa(10000+i)
	network.Listen(address)
	id, _ := contact.NewKademliaIDFromAddress(address)
	me := contact.NewContact(id, address)
	rpc := rpc.CreateRpc(receiver, network, me)
	node := NewKademliaNode(config, rpc)

	mock.nodes = append(mock.nodes, node)

	fmt.Printf("Node %s with address: %s created\n", address, id.String())

	return node
}

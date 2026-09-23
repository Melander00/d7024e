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
	Simulation *network.Simulation
	Config     KademliaConfig
	Nodes      []*Kademlia
}

func MockKademlia(nrNodes int) *Mock {
	fmt.Printf("Mocking kademlia %d nodes\n", nrNodes)

	simulation := network.NewSimulation(0, 0, 1)

	config := KademliaConfig{
		Alpha:   3,
		K:       10,
		Timeout: 5 * time.Second,
		Retries: 5,
	}

	mock := &Mock{
		Simulation: simulation,
		Config:     config,
	}

	boot_node := mock.CreateNode(0, simulation, config)

	for i := 1; i < nrNodes; i++ {
		node := mock.CreateNode(i, simulation, config)
		go node.Join(boot_node.Me)
	}

	// time.Sleep(1 * time.Second)

	// nodes := boot_node.Routing.FindClosestContacts(boot_node.Me.ID, nrNodes)
	// fmt.Printf("Boot has %d nodes in routing table\n", len(nodes))
	// nodes = mock.Nodes[1].Routing.FindClosestContacts(boot_node.Me.ID, nrNodes)
	// fmt.Printf("Node-10001 has %d nodes\n", len(nodes))
	// for _, node := range nodes {

	// 	fmt.Printf("\t%s %s \n", node.Address, node.ID)
	// }

	return mock
}

func (mock *Mock) CreateNode(i int, simulation *network.Simulation, config KademliaConfig) *Kademlia {
	receiver := network.NewChannelNetworkReceiver(5)
	network := simulation.NewSimulatedNetwork(receiver)
	address := "127.0.0.1" + ":" + strconv.Itoa(10000+i)
	network.Listen(address)
	id, _ := contact.NewKademliaIDFromAddress(address)
	me := contact.NewContact(id, address)
	rpc := rpc.CreateRpc(receiver, network, me)
	rpc.SetDataNetwork(simulation.NewSimulatedDataNetwork())
	if err := rpc.StartDataPlane(); err != nil {
		panic(err)
	}
	node := NewKademliaNode(config, rpc)

	mock.Nodes = append(mock.Nodes, node)

	// fmt.Printf("Node %s with address: %s created\n", address, id.String())

	return node
}

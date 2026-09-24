//go:build experiment

package main

import (
	"d7024e/internal/kademlia"
	"d7024e/internal/kademlia/contact"
	"d7024e/internal/kademlia/rpc"
	"d7024e/pkg/logger"
	"d7024e/pkg/network"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

type experiment struct {
	rand   *rand.Rand
	config kademlia.KademliaConfig
	sim    *network.Simulation
}

func main() {
	logger := logger.NewFileLogger("./data/scalability.txt")
	fmt.Println("Starting scalability experiment.")
	fmt.Println("")
	for i, n := range []int{10, 50, 100, 500, 1000, 2000} {
		for _, seed := range []int64{1, 2, 3, 4, 5} {
			fmt.Printf("nodes=%d \t seed=%d\n", n, seed)
			logger.Log(fmt.Sprintf("\n# Experiment nr: %d \t nodes=%d \t seed=%d \t lookups=%d\n", i, n, seed, 10))
			runScalabilityExperiment(n, seed, 10, logger)
		}
	}
}

func runScalabilityExperiment(nrNodes int, seed int64, nrLookups int, logger logger.Logger) {
	// Parameters
	latency := 0.0
	packetLoss := 0.0

	// Setup deterministic randomize
	rand := rand.New(rand.NewSource(seed))

	// Create simulation & config
	sim := network.NewSimulation(latency, packetLoss, seed)
	config := kademlia.KademliaConfig{
		Alpha:                 3,
		K:                     10,
		Timeout:               5 * time.Second,
		Retries:               5,
		BucketRefreshInterval: 1 * time.Hour,
	}

	exp := &experiment{
		rand:   rand,
		config: config,
		sim:    sim,
	}

	// Create boot node
	boot := exp.createNode(0)
	boot.Logger = logger

	// Create N nodes that join using boot node
	for i := 1; i < nrNodes; i++ {
		node := exp.createNode(i)
		go node.Join(boot.Me)
	}

	// Wait for eventual consistency
	time.Sleep(300 * time.Millisecond)

	for _ = range nrLookups {
		targetID := exp.getRandomID()
		targetContact := &contact.Contact{
			ID: targetID,
		}

		boot.LookupContact(targetContact)
	}
}

func (exp *experiment) createNode(i int) *kademlia.Kademlia {
	receiver := network.NewChannelNetworkReceiver(5)
	network := exp.sim.NewSimulatedNetwork(receiver)

	address := exp.randomizeIP() + ":" + strconv.Itoa(10000+i)

	network.Listen(address)
	id, _ := contact.NewKademliaIDFromAddress(address)
	me := contact.NewContact(id, address)
	rpc := rpc.CreateRpc(receiver, network, me)
	rpc.SetDataNetwork(exp.sim.NewSimulatedDataNetwork())
	if err := rpc.StartDataPlane(); err != nil {
		panic(err)
	}
	node := kademlia.NewKademliaNode(exp.config, rpc)

	return node
}

func (exp *experiment) randomizeIP() string {
	return fmt.Sprintf(
		"%d.%d.%d.%d",
		exp.rand.Intn(256),
		exp.rand.Intn(256),
		exp.rand.Intn(256),
		exp.rand.Intn(256),
	)
}

func (exp *experiment) getRandomID() *contact.KademliaID {
	newKademliaID := contact.KademliaID{}
	if _, err := exp.rand.Read(newKademliaID[:]); err != nil {
		panic(fmt.Errorf("kademlia: failed to generate random ID: %w", err))
	}

	return &newKademliaID
}

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

const (
	networkSize = 10
	lookups     = 10
)

var packetLossProbabilities = []float64{0, 0.05, 0.1, 0.2, 0.3, 0.4, 0.5}
var experimentSeeds = []int64{1, 2, 3, 4, 5}

type experiment struct {
	rand   *rand.Rand
	config kademlia.KademliaConfig
	sim    *network.Simulation
	nodes  []*kademlia.Kademlia
}

func main() {
	log := logger.NewFileLogger("./data/packet_loss.txt")

	for experimentNumber, packetLoss := range packetLossProbabilities {
		for _, seed := range experimentSeeds {
			fmt.Printf("packet_loss=%.2f seed=%d\n", packetLoss, seed)
			log.Log(fmt.Sprintf(
				"\n# Experiment nr: %d nodes=%d packet_loss=%.2f seed=%d lookups=%d\n",
				experimentNumber, networkSize, packetLoss, seed, lookups,
			))
			runExperiment(networkSize, packetLoss, seed, lookups, log)
		}
	}
}

func runExperiment(nrNodes int, packetLoss float64, seed int64, nrLookups int, log logger.Logger) {
	random := rand.New(rand.NewSource(seed))
	simulation := network.NewSimulation(0, packetLoss, seed)
	config := kademlia.KademliaConfig{
		Alpha:                 3,
		K:                     10,
		Timeout:               100 * time.Millisecond,
		Retries:               3,
		BucketRefreshInterval: time.Hour,
	}

	exp := &experiment{
		rand:   random,
		config: config,
		sim:    simulation,
		nodes:  make([]*kademlia.Kademlia, 0, nrNodes),
	}

	for i := 0; i < nrNodes; i++ {
		exp.nodes = append(exp.nodes, exp.createNode(i))
	}

	for i := 1; i < nrNodes; i++ {
		exp.nodes[i].Join(exp.nodes[0].Me)
	}

	values := make([][]byte, nrLookups)
	keys := make([]*contact.KademliaID, nrLookups)
	for i := range values {
		values[i] = randomValue(random)
		keys[i] = contact.NewKademliaIDFromData(values[i])
		if err := exp.nodes[0].Datastore.Put(keys[i], values[i]); err != nil {
			panic(err)
		}
	}

	for _, node := range exp.nodes {
		node.Logger = log
	}

	for i := 0; i < nrLookups; i++ {
		requester := exp.nodes[(i%(nrNodes-1))+1]
		value, err := requester.LookupData(keys[i].String())
		success := err == nil && string(value) == string(values[i])
		log.Log(fmt.Sprintf(
			"%s lookup_result %d success=%t error=%q\n",
			requester.Me.Address, i, success, errorString(err),
		))
	}
}

func (exp *experiment) createNode(i int) *kademlia.Kademlia {
	receiver := network.NewChannelNetworkReceiver(5)
	net := exp.sim.NewSimulatedNetwork(receiver)
	address := exp.randomIP() + ":" + strconv.Itoa(10000+i)
	if err := net.Listen(address); err != nil {
		panic(err)
	}

	id, _ := contact.NewKademliaIDFromAddress(address)
	me := contact.NewContact(id, address)
	rpcClient := rpc.CreateRpc(receiver, net, me)
	rpcClient.SetDataNetwork(exp.sim.NewSimulatedDataNetwork())
	if err := rpcClient.StartDataPlane(); err != nil {
		panic(err)
	}

	return kademlia.NewKademliaNode(exp.config, rpcClient)
}

func (exp *experiment) randomIP() string {
	return fmt.Sprintf(
		"%d.%d.%d.%d",
		exp.rand.Intn(256), exp.rand.Intn(256), exp.rand.Intn(256), exp.rand.Intn(256),
	)
}

func randomValue(random *rand.Rand) []byte {
	value := make([]byte, 32)
	if _, err := random.Read(value); err != nil {
		panic(err)
	}
	return value
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

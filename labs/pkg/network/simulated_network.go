package network

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

type SimulatedNetwork struct {
	receiver   NetworkReceiver
	simulation *Simulation
	Address    string
}

type Simulation struct {
	latency    float64
	packetLoss float64
	mu         sync.RWMutex

	nodes        map[string]*SimulatedNetwork
	dataHandlers map[string]DataHandler

	rand *rand.Rand
}

func NewSimulation(latency float64, packetLoss float64, randSeed int64) *Simulation {
	return &Simulation{
		nodes:        make(map[string]*SimulatedNetwork),
		dataHandlers: make(map[string]DataHandler),
		latency:      latency,
		packetLoss:   packetLoss,
		rand:         rand.New(rand.NewSource(randSeed)),
	}
}

func (simulation *Simulation) NewSimulatedNetwork(receiver NetworkReceiver) *SimulatedNetwork {
	return &SimulatedNetwork{
		receiver:   receiver,
		simulation: simulation,
	}
}

func (net *SimulatedNetwork) Listen(address string) error {
	net.Address = address

	net.simulation.mu.Lock()

	net.simulation.nodes[net.Address] = net

	net.simulation.mu.Unlock()

	return nil
}

func (net *SimulatedNetwork) Send(to string, bytes []byte) error {
	net.simulation.mu.RLock()

	dest, exists := net.simulation.nodes[to]

	net.simulation.mu.RUnlock()

	if !exists {
		return errors.New("Target doesn't exist")
	}

	go func() {
		// First we sleep to simulate latency
		net.simulation.mu.RLock()
		latency := time.Duration(net.simulation.latency * float64(time.Millisecond))
		net.simulation.mu.RUnlock()

		time.Sleep(latency)

		// Then we randomize if we should drop the packet
		net.simulation.mu.Lock()
		loss := net.simulation.rand.Float64()

		pLoss := net.simulation.packetLoss
		net.simulation.mu.Unlock()

		if loss < pLoss {
			// Drop packet
			return
		}

		dest.receiver.OnData(bytes)
	}()

	return nil
}

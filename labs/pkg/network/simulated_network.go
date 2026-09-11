package network

import (
	"errors"
	"sync"
)

type SimulatedNetwork struct {
	receiver   NetworkReceiver
	simulation *Simulation
	Address    Address
}

type Simulation struct {
	mu    sync.RWMutex
	nodes map[Address]*SimulatedNetwork
}

func NewSimulation() *Simulation {
	return &Simulation{
		nodes: make(map[Address]*SimulatedNetwork),
	}
}

func (simulation *Simulation) NewSimulatedNetwork(receiver NetworkReceiver) *SimulatedNetwork {
	return &SimulatedNetwork{
		receiver:   receiver,
		simulation: simulation,
	}
}

func (net *SimulatedNetwork) Listen(address Address) error {
	net.Address = address

	net.simulation.mu.Lock()

	net.simulation.nodes[net.Address] = net

	net.simulation.mu.Unlock()

	return nil
}

func (net *SimulatedNetwork) Send(to Address, bytes []byte) error {
	net.simulation.mu.RLock()

	dest, exists := net.simulation.nodes[to]

	net.simulation.mu.RUnlock()

	if !exists {
		return errors.New("Target doesn't exist")
	}

	go dest.receiver.OnData(bytes) // goroutine so it simulates async behaviour.

	return nil
}

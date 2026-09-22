package network

import "errors"

type SimulatedDataNetwork struct {
	simulation *Simulation
}

func (simulation *Simulation) NewSimulatedDataNetwork() *SimulatedDataNetwork {
	return &SimulatedDataNetwork{
		simulation: simulation,
	}
}

// Listen registers a data handler for the given address in the simulated network.
func (n *SimulatedDataNetwork) Listen(
	address string,
	handler DataHandler,
) error {

	n.simulation.mu.Lock()
	defer n.simulation.mu.Unlock()

	n.simulation.dataHandlers[address] = handler

	return nil
}

// Request sends data to the specified address in the simulated network and returns the response.
func (n *SimulatedDataNetwork) Request(
	to string,
	data []byte,
) ([]byte, error) {

	n.simulation.mu.RLock()
	handler, exists := n.simulation.dataHandlers[to]
	n.simulation.mu.RUnlock()

	if !exists {
		return nil, errors.New("data target does not exist")
	}

	// Optional: simulate latency here.
	// Do NOT apply UDP packet loss.

	return handler(data), nil
}

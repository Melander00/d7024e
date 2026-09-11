package network

import (
	"net"
)

type UdpNetwork struct {
	conn     *net.UDPConn
	receiver NetworkReceiver
}

func (net *UdpNetwork) Listen(address Address) error {
	// Create new UDP connection

	// Create listener

	// Start goroutine that constantly receives data
	// It should forward the data to RPC

	go net.receive()

	return nil
}

func (net *UdpNetwork) Send(to string, bytes []byte) error {
	// Translate Contact to UDP address
	// Send message via UDP

	return nil
}

func (net *UdpNetwork) receive() {
	// While True
	// Read bytestream
	// net.receiver.OnData(data)
}

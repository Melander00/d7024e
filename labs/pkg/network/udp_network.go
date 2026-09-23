package network

import (
	"net"
)

type UdpNetwork struct {
	conn     *net.UDPConn
	receiver NetworkReceiver
	Address  string
}

func NewUdpNetwork(receiver NetworkReceiver) *UdpNetwork {
	return &UdpNetwork{
		receiver: receiver,
	}
}

func (n *UdpNetwork) Listen(address string) error {
	n.Address = address

	addr, err := net.ResolveUDPAddr("udp", address)

	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)

	if err != nil {
		return err
	}

	n.conn = conn

	go n.receive()

	return nil
}

func (n *UdpNetwork) Send(to string, bytes []byte) error {
	addr, err := net.ResolveUDPAddr("udp", to)
	if err != nil {
		return err
	}

	_, writeErr := n.conn.WriteToUDP(bytes, addr)
	if writeErr != nil {
		return writeErr
	}

	return nil
}

func (n *UdpNetwork) receive() {
	buf := make([]byte, 4096)
	defer n.conn.Close()
	for {
		nr, _, err := n.conn.ReadFromUDP(buf)

		if err != nil {
			return
		}

		// Copy so the buffer doesnt get overwritten due to race condition.
		data := append([]byte(nil), buf[:nr]...)

		go n.receiver.OnData(data)
	}
}

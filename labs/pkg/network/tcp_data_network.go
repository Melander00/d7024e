package network

import (
	"encoding/binary"
	"io"
	"net"
)

type TCPDataNetwork struct {
	handler DataHandler
}

func NewTCPDataNetwork() *TCPDataNetwork {
	return &TCPDataNetwork{}
}

// Listen starts a TCP server on the given address and uses the provided handler to process incoming data.
func (n *TCPDataNetwork) Listen(
	address string,
	handler DataHandler,
) error {

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	n.handler = handler

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}

			go func() {
				defer conn.Close()

				request, err := readFrame(conn)
				if err != nil {
					return
				}

				response := handler(request)

				_ = writeFrame(conn, response)
			}()
		}
	}()

	return nil
}

// Request sends data to the specified address over TCP and returns the response.
func (n *TCPDataNetwork) Request(
	to string,
	data []byte,
) ([]byte, error) {

	conn, err := net.Dial("tcp", to)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := writeFrame(conn, data); err != nil {
		return nil, err
	}

	return readFrame(conn)
}

func writeFrame(conn net.Conn, data []byte) error {
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(data)))

	if err := writeAll(conn, length); err != nil {
		return err
	}

	return writeAll(conn, data)
}

func writeAll(conn net.Conn, data []byte) error {
	for len(data) > 0 {
		n, err := conn.Write(data)
		if err != nil {
			return err
		}

		if n == 0 {
			return io.ErrUnexpectedEOF
		}

		data = data[n:]
	}

	return nil
}

func readFrame(conn net.Conn) ([]byte, error) {
	length := make([]byte, 4)
	if _, err := io.ReadFull(conn, length); err != nil {
		return nil, err
	}

	payload := make([]byte, binary.BigEndian.Uint32(length))
	if _, err := io.ReadFull(conn, payload); err != nil {
		return nil, err
	}

	return payload, nil
}

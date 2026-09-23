package network

import (
	"bytes"
	"testing"
	"time"
)

type testReceiver struct {
	data chan []byte
}

func (r *testReceiver) OnData(data []byte) {
	r.data <- data
}

func (r *testReceiver) Read(out chan []byte) {
	// do nothing for these tests
}

func TestUdpNetworkSendReceive(t *testing.T) {
	receiver := &testReceiver{
		data: make(chan []byte, 1),
	}

	network := NewUdpNetwork(receiver)

	// :0 asks the OS to choose an available UDP port.
	if err := network.Listen("127.0.0.1:0"); err != nil {
		t.Fatal(err)
	}
	defer network.conn.Close()

	// The actual port is assigned by the OS.
	address := network.conn.LocalAddr().String()

	message := []byte("hello UDP")

	if err := network.Send(address, message); err != nil {
		t.Fatal(err)
	}

	select {
	case received := <-receiver.data:
		if !bytes.Equal(received, message) {
			t.Fatalf("received %q, want %q", received, message)
		}

	case <-time.After(time.Second):
		t.Fatal("timed out waiting for UDP message")
	}
}

func TestUdpNetworkCommunication(t *testing.T) {
	receiver := &testReceiver{
		data: make(chan []byte, 1),
	}

	a := NewUdpNetwork(&testReceiver{
		data: make(chan []byte, 1),
	})

	b := NewUdpNetwork(receiver)

	if err := a.Listen("127.0.0.1:0"); err != nil {
		t.Fatal(err)
	}
	if err := b.Listen("127.0.0.1:0"); err != nil {
		t.Fatal(err)
	}

	defer a.conn.Close()
	defer b.conn.Close()

	message := []byte("hello from A")

	if err := a.Send(b.conn.LocalAddr().String(), message); err != nil {
		t.Fatal(err)
	}

	select {
	case received := <-receiver.data:
		if !bytes.Equal(received, message) {
			t.Fatalf("received %q, want %q", received, message)
		}

	case <-time.After(time.Second):
		t.Fatal("timed out waiting for message")
	}
}

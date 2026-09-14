package network

import (
	"bytes"
	"testing"
	"time"
)

// TestSimulatedNetworkSend verifies that data sent through the simulated network
// is delivered to the receiver registered at the destination address.
func TestSimulatedNetworkSend(t *testing.T) {
	simulation := NewSimulation()

	receiverA := NewChannelNetworkReceiver(1)
	receiverB := NewChannelNetworkReceiver(1)

	networkA := simulation.NewSimulatedNetwork(receiverA)
	networkB := simulation.NewSimulatedNetwork(receiverB)

	if err := networkA.Listen("node-a"); err != nil {
		t.Fatal(err)
	}

	if err := networkB.Listen("node-b"); err != nil {
		t.Fatal(err)
	}

	received := make(chan []byte, 1)
	go receiverB.Read(received)

	expected := []byte("hello")

	if err := networkA.Send("node-b", expected); err != nil {
		t.Fatal(err)
	}

	select {
	case actual := <-received:
		if !bytes.Equal(actual, expected) {
			t.Fatalf("expected %q, got %q", expected, actual)
		}

	case <-time.After(time.Second):
		t.Fatal("timed out waiting for receiver B")
	}
}

// TestSimulatedNetworkSendUnknownAddress verifies that sending data to an
// address that is not registered in the simulation returns an error.
func TestSimulatedNetworkSendUnknownAddress(t *testing.T) {
	simulation := NewSimulation()
	receiverA := NewChannelNetworkReceiver(1)
	networkA := simulation.NewSimulatedNetwork(receiverA)

	if err := networkA.Listen("node-a"); err != nil {
		t.Fatal(err)
	}

	if err := networkA.Send("node-b", []byte("hello")); err == nil {
		t.Fatal("expected error when sending to unknown address, got nil")
	}
}

package rpc

import (
	"bytes"
	"d7024e/internal/kademlia/contact"
	"d7024e/pkg/network"
	"testing"
)

// TestRPCPing verifies that one RPC node can send a PING request through the
// simulated network and receive the corresponding PONG response.
func TestRPCPing(t *testing.T) {
	simulation := network.NewSimulation(0, 0, 1)

	receiverA := network.NewChannelNetworkReceiver(1)
	receiverB := network.NewChannelNetworkReceiver(1)

	networkA := simulation.NewSimulatedNetwork(receiverA)
	networkB := simulation.NewSimulatedNetwork(receiverB)

	addressA := "127.0.0.1:10001"
	addressB := "127.0.0.1:10002"

	if err := networkA.Listen(addressA); err != nil {
		t.Fatal(err)
	}
	if err := networkB.Listen(addressB); err != nil {
		t.Fatal(err)
	}

	idA, err := contact.NewKademliaIDFromAddress(addressA)
	if err != nil {
		t.Fatal(err)
	}

	idB, err := contact.NewKademliaIDFromAddress(addressB)
	if err != nil {
		t.Fatal(err)
	}

	contactA := contact.NewContact(idA, addressA)
	contactB := contact.NewContact(idB, addressB)

	rpcA := CreateRpc(receiverA, networkA, contactA)
	rpcB := CreateRpc(receiverB, networkB, contactB)

	rpcB.SetPingHandler(func(client contact.Contact) {
		// nothing required
	})

	response, err := rpcA.Ping(contactB)
	if err != nil {
		t.Fatal(err)
	}

	if response.Type != PONG {
		t.Fatalf("expected PONG, got %s", response.Type)
	}
}

// TestRPCFindNode verifies that one RPC node can send a FIND_NODE request
// through the simulated network and receive the contacts returned by the
// destination node's FindNode handler.
func TestRPCFindNode(t *testing.T) {
	simulation := network.NewSimulation(0, 0, 1)

	receiverA := network.NewChannelNetworkReceiver(1)
	receiverB := network.NewChannelNetworkReceiver(1)

	networkA := simulation.NewSimulatedNetwork(receiverA)
	networkB := simulation.NewSimulatedNetwork(receiverB)

	addressA := "127.0.0.1:10001"
	addressB := "127.0.0.1:10002"

	if err := networkA.Listen(addressA); err != nil {
		t.Fatal(err)
	}
	if err := networkB.Listen(addressB); err != nil {
		t.Fatal(err)
	}

	idA, err := contact.NewKademliaIDFromAddress(addressA)
	if err != nil {
		t.Fatal(err)
	}

	idB, err := contact.NewKademliaIDFromAddress(addressB)
	if err != nil {
		t.Fatal(err)
	}

	contactA := contact.NewContact(idA, addressA)
	contactB := contact.NewContact(idB, addressB)

	rpcA := CreateRpc(receiverA, networkA, contactA)
	rpcB := CreateRpc(receiverB, networkB, contactB)

	rpcB.SetFindNodeHandler(func(client contact.Contact, hash string) ([]contact.Contact, error) {
		return []contact.Contact{contactA}, nil
	})

	response, err := rpcA.FindNode(
		contactB,
		idA.String(),
	)
	if err != nil {
		t.Fatal(err)
	}

	if response.Type != FIND_NODE_RESPONSE {
		t.Fatalf("expected FIND_NODE_RESPONSE, got %s", response.Type)
	}

	if len(response.Value.Nodes) != 1 {
		t.Fatalf("expected 1 contact, got %d", len(response.Value.Nodes))
	}

	if !response.Value.Nodes[0].ID.Equals(
		idA,
	) {
		t.Fatal("expected returned contact to be contactA")
	}
}

// TestRPCFindValue verifies that one RPC node can send a FIND_VALUE request
// through the simulated network and receive the value returned by the
// destination node's FindValue handler.
func TestRPCFindValue(t *testing.T) {
	simulation := network.NewSimulation(0, 0, 1)

	receiverA := network.NewChannelNetworkReceiver(1)
	receiverB := network.NewChannelNetworkReceiver(1)

	networkA := simulation.NewSimulatedNetwork(receiverA)
	networkB := simulation.NewSimulatedNetwork(receiverB)

	addressA := "127.0.0.1:10001"
	addressB := "127.0.0.1:10002"

	if err := networkA.Listen(addressA); err != nil {
		t.Fatal(err)
	}
	if err := networkB.Listen(addressB); err != nil {
		t.Fatal(err)
	}

	idA, err := contact.NewKademliaIDFromAddress(addressA)
	if err != nil {
		t.Fatal(err)
	}

	idB, err := contact.NewKademliaIDFromAddress(addressB)
	if err != nil {
		t.Fatal(err)
	}

	contactA := contact.NewContact(idA, addressA)
	contactB := contact.NewContact(idB, addressB)

	rpcA := CreateRpc(receiverA, networkA, contactA)
	rpcB := CreateRpc(receiverB, networkB, contactB)
	rpcA.SetDataNetwork(simulation.NewSimulatedDataNetwork())
	rpcB.SetDataNetwork(simulation.NewSimulatedDataNetwork())
	if err := rpcA.StartDataPlane(); err != nil {
		t.Fatal(err)
	}
	if err := rpcB.StartDataPlane(); err != nil {
		t.Fatal(err)
	}

	expected := []byte("hello")
	target := contact.NewKademliaIDFromData(expected)

	rpcB.SetFindValueHandler(
		func(client contact.Contact, hash string) ([]contact.Contact, []byte, bool, error) {
			return nil, expected, true, nil
		},
	)

	response, err := rpcA.FindValue(contactB, target.String())
	if err != nil {
		t.Fatal(err)
	}

	if response.Type != FIND_VALUE_RESPONSE {
		t.Fatalf("expected FIND_VALUE_RESPONSE, got %s", response.Type)
	}

	if !bytes.Equal(response.Value.Value, expected) {
		t.Fatalf("expected %q, got %q", expected, response.Value.Value)
	}
}

// TestRPCStore verifies that one RPC node can send a STORE request through
// the simulated network and receive a successful STORE response from the
// destination node's Store handler.
func TestRPCStore(t *testing.T) {
	simulation := network.NewSimulation(0, 0, 1)

	receiverA := network.NewChannelNetworkReceiver(1)
	receiverB := network.NewChannelNetworkReceiver(1)

	networkA := simulation.NewSimulatedNetwork(receiverA)
	networkB := simulation.NewSimulatedNetwork(receiverB)

	addressA := "127.0.0.1:10001"
	addressB := "127.0.0.1:10002"

	if err := networkA.Listen(addressA); err != nil {
		t.Fatal(err)
	}
	if err := networkB.Listen(addressB); err != nil {
		t.Fatal(err)
	}

	idA, err := contact.NewKademliaIDFromAddress(addressA)
	if err != nil {
		t.Fatal(err)
	}

	idB, err := contact.NewKademliaIDFromAddress(addressB)
	if err != nil {
		t.Fatal(err)
	}

	contactA := contact.NewContact(idA, addressA)
	contactB := contact.NewContact(idB, addressB)

	rpcA := CreateRpc(receiverA, networkA, contactA)
	rpcB := CreateRpc(receiverB, networkB, contactB)
	rpcA.SetDataNetwork(simulation.NewSimulatedDataNetwork())
	rpcB.SetDataNetwork(simulation.NewSimulatedDataNetwork())
	if err := rpcA.StartDataPlane(); err != nil {
		t.Fatal(err)
	}
	if err := rpcB.StartDataPlane(); err != nil {
		t.Fatal(err)
	}

	key := contact.NewRandomKademliaID().String()
	value := []byte("hello")

	rpcB.SetStoreHandler(
		func(client contact.Contact, receivedKey string, receivedValue []byte) bool {
			if receivedKey != key {
				return false
			}

			if !bytes.Equal(receivedValue, value) {
				return false
			}

			return true
		},
	)

	response, err := rpcA.Store(contactB, key, value)
	if err != nil {
		t.Fatal(err)
	}

	if response.Type != STORE_RESPONSE {
		t.Fatalf("expected STORE_RESPONSE, got %s", response.Type)
	}

	if !response.Value.Success {
		t.Fatal("expected store operation to succeed")
	}
}

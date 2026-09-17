package kademlia

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"d7024e/internal/kademlia/contact"
	"d7024e/internal/kademlia/rpc"
	"d7024e/pkg/network"
)

// createTestNode creates a fully wired Kademlia node attached to the simulation.
func createTestNode(t *testing.T, simulation *network.Simulation, address string) *Kademlia {
	t.Helper()

	receiver := network.NewChannelNetworkReceiver(10)
	net := simulation.NewSimulatedNetwork(receiver)

	if err := net.Listen(address); err != nil {
		t.Fatalf("failed to listen on %s: %v", address, err)
	}

	id, err := contact.NewKademliaIDFromAddress(address)
	if err != nil {
		t.Fatalf("failed to generate Kademlia ID for %s: %v", address, err)
	}

	me := contact.NewContact(id, address)
	rpcNode := rpc.CreateRpc(receiver, net, me)

	config := KademliaConfig{
		Alpha:   3,
		K:       10,
		Timeout: 500 * time.Millisecond,
		Retries: 2,
	}

	return NewKademliaNode(config, rpcNode)
}

// TestKademliaJoinAndLookupContact verifies that after node B joins via bootstrap node A,
// node B can locate node A via LookupContact.
func TestKademliaJoinAndLookupContact(t *testing.T) {
	simulation := network.NewSimulation()

	nodeA := createTestNode(t, simulation, "127.0.0.1:10001")
	nodeB := createTestNode(t, simulation, "127.0.0.1:10002")

	nodeB.Join(nodeA.Me)

	results := nodeB.LookupContact(&nodeA.Me)

	found := false
	for _, c := range results {
		if c.ID.Equals(nodeA.Me.ID) {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected node A (%s) in lookup results, got %v", nodeA.Me.ID.String(), results)
	}
}

// TestKademliaLookupContactAcrossMultipleNodes tests iterative lookup across multiple hops:
// node A joins B, node C joins B, and node A discovers node C via LookupContact.
func TestKademliaLookupContactAcrossMultipleNodes(t *testing.T) {
	simulation := network.NewSimulation()

	nodeA := createTestNode(t, simulation, "127.0.0.1:10001")
	nodeB := createTestNode(t, simulation, "127.0.0.1:10002")
	nodeC := createTestNode(t, simulation, "127.0.0.1:10003")

	nodeA.Join(nodeB.Me)
	nodeC.Join(nodeB.Me)

	results := nodeA.LookupContact(&nodeC.Me)

	found := false
	for _, c := range results {
		if c.ID.Equals(nodeC.Me.ID) {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("expected node A to discover node C (%s) through node B, got %v", nodeC.Me.ID.String(), results)
	}
}

// TestKademliaLookupDataFound tests remote data retrieval when node A has stored data
// and node B requests it via LookupData.
func TestKademliaLookupDataFound(t *testing.T) {
	simulation := network.NewSimulation()

	nodeA := createTestNode(t, simulation, "127.0.0.1:10001")
	nodeB := createTestNode(t, simulation, "127.0.0.1:10002")

	nodeB.Join(nodeA.Me)

	expectedData := []byte("hello")
	key := contact.NewRandomKademliaID()

	if err := nodeA.Datastore.Put(key, expectedData); err != nil {
		t.Fatal(err)
	}

	result := nodeB.LookupData(key.String())

	if !bytes.Equal(result, expectedData) {
		t.Fatalf("expected data %q, got %q", expectedData, result)
	}
}

// TestKademliaLookupDataMissing checks that LookupData returns nil/empty
// when no node in the network possesses the requested key.
func TestKademliaLookupDataMissing(t *testing.T) {
	simulation := network.NewSimulation()

	nodeA := createTestNode(t, simulation, "127.0.0.1:10001")
	nodeB := createTestNode(t, simulation, "127.0.0.1:10002")

	nodeB.Join(nodeA.Me)

	missingKey := contact.NewRandomKademliaID()

	result := nodeB.LookupData(missingKey.String())

	if len(result) != 0 {
		t.Fatalf("expected empty result for missing key, got %q", result)
	}
}

// TestKademliaStoreAndLookupData tests the full Store and remote Lookup flow.

func TestKademliaStoreAndLookupData(t *testing.T) {
	simulation := network.NewSimulation()

	nodeA := createTestNode(t, simulation, "127.0.0.1:10001")
	nodeB := createTestNode(t, simulation, "127.0.0.1:10002")
	nodeC := createTestNode(t, simulation, "127.0.0.1:10003")

	nodeB.Join(nodeA.Me)
	nodeC.Join(nodeB.Me)

	data := []byte("hello")

	nodeA.Store(data)

	hashBytes := sha256.Sum256(data)
	hash := hex.EncodeToString(hashBytes[:])

	result := nodeC.LookupData(hash)

	if !bytes.Equal(result, data) {
		t.Fatalf("expected stored data %q, got %q", data, result)
	}
}

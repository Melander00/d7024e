package kademlia

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	"d7024e/internal/kademlia/contact"
	"d7024e/pkg/network"
)

func TestReplicateOnceDiscoversClosestNodes(t *testing.T) {
	simulation := network.NewSimulation(0, 0, 1)
	value := []byte("replicate near this key")
	key := contact.NewKademliaIDFromData(value)
	nodes := make([]*Kademlia, 5)
	for i := range nodes {
		nodes[i] = createTestNode(t, simulation, fmt.Sprintf("127.0.0.1:%d", 12000+i))
	}
	// Order by XOR distance: two intended recipients, one farther candidate,
	// then the relay and holder. This keeps the expected placement deterministic.
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID.CalcDistance(key).Less(nodes[j].ID.CalcDistance(key))
	})

	// The holder knows only the relay. A fresh iterative lookup must discover
	// the two closest nodes through that relay, rather than reuse local contacts.
	holder, relay := nodes[4], nodes[3]
	holder.Config.K = 2
	holder.Routing.AddContact(relay.Me)
	for _, node := range nodes[:3] {
		relay.Routing.AddContact(node.Me)
	}
	if err := holder.Datastore.Put(key, value); err != nil {
		t.Fatal(err)
	}

	holder.replicateOnce()

	for _, node := range []*Kademlia{nodes[0], nodes[1], holder} {
		assertReplicationValue(t, node, key, value)
	}
	for _, node := range []*Kademlia{nodes[2], relay} {
		if _, exists := node.Datastore.Get(key); exists {
			t.Fatalf("replicated to %s outside the closest K nodes", node.Me.Address)
		}
	}
}

func TestReplicationLoopSurvivesHolderFailures(t *testing.T) {
	simulation := network.NewSimulation(0, 0, 1)
	original := createTestNode(t, simulation, "127.0.0.1:12100")
	survivor := createTestNode(t, simulation, "127.0.0.1:12101")
	original.Routing.AddContact(survivor.Me)
	value := []byte("survive successive holder failures")
	key := contact.NewKademliaIDFromData(value)
	original.Store(value)
	assertReplicationValue(t, survivor, key, value)
	makeReplicationNodeUnavailable(t, simulation, original)

	// Only the replica runs maintenance; the original publisher is unavailable.
	first := createTestNode(t, simulation, "127.0.0.1:12102")
	survivor.Routing.AddContact(first.Me)
	if _, exists := first.Datastore.Get(key); exists {
		t.Fatal("replacement already holds the value before replication starts")
	}
	survivor.Config.ReplicationInterval = 20 * time.Millisecond
	survivor.Rpc.SetTimeout(20 * time.Millisecond)
	survivor.Rpc.SetRetries(1)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		survivor.runReplicationLoop(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("replication loop did not stop after cancellation")
		}
	})
	waitForReplicationValue(t, first, key, value)

	// Introduce a new destination after the first copy arrives. Copying to it must
	// happen on a later tick, with another lookup of the key.
	second := createTestNode(t, simulation, "127.0.0.1:12103")
	survivor.Routing.AddContact(second.Me)
	waitForReplicationValue(t, second, key, value)
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("replication loop did not stop after cancellation")
	}

	// Make every holder except the newest replica unavailable. Retrieval must use
	// a copy created by periodic replication, not an original surviving copy.
	makeReplicationNodeUnavailable(t, simulation, survivor)
	makeReplicationNodeUnavailable(t, simulation, first)
	reader := createTestNode(t, simulation, "127.0.0.1:12104")
	reader.Routing.AddContact(second.Me)
	got, err := reader.LookupData(key.String())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, value) {
		t.Fatalf("lookup after holder failures returned %q, want %q", got, value)
	}
}

func assertReplicationValue(t *testing.T, node *Kademlia, key *contact.KademliaID, want []byte) {
	t.Helper()
	got, exists := node.Datastore.Get(key)
	if !exists || !bytes.Equal(got, want) {
		t.Fatalf("node %s: value = %q, exists = %v; want %q", node.Me.Address, got, exists, want)
	}
}

func waitForReplicationValue(t *testing.T, node *Kademlia, key *contact.KademliaID, want []byte) {
	t.Helper()
	deadline := time.NewTimer(2 * time.Second)
	defer deadline.Stop()
	poll := time.NewTicker(5 * time.Millisecond)
	defer poll.Stop()
	for {
		if _, exists := node.Datastore.Get(key); exists {
			assertReplicationValue(t, node, key, want)
			return
		}
		select {
		case <-poll.C:
		case <-deadline.C:
			t.Fatalf("timed out waiting for replication to %s", node.Me.Address)
		}
	}
}

// The simulator has no removal API. Model node failure by replacing its
// registered endpoints: control requests time out and data requests fail.
// The node and its datastore still exist, but are unreachable through the network.
// Any replication loop on this node must be stopped by the caller first.
func makeReplicationNodeUnavailable(t *testing.T, simulation *network.Simulation, node *Kademlia) {
	t.Helper()
	control := simulation.NewSimulatedNetwork(replicationOfflineReceiver{})
	if err := control.Listen(node.Me.Address); err != nil {
		t.Fatal(err)
	}
	if err := simulation.NewSimulatedDataNetwork().Listen(node.Me.Address, func([]byte) []byte {
		return []byte(`{"OK":false,"Error":"node offline"}`)
	}); err != nil {
		t.Fatal(err)
	}
}

type replicationOfflineReceiver struct{}

func (replicationOfflineReceiver) OnData([]byte)    {}
func (replicationOfflineReceiver) Read(chan []byte) {}

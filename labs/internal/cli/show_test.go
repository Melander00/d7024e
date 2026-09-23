package cli

import (
	"bytes"
	"d7024e/internal/kademlia"
	"d7024e/internal/kademlia/contact"
	"d7024e/pkg/network"
	"io"
	"os"
	"strings"
	"time"
)

func captureOutput(t *T, fn func()) string {
	t.Helper()

	old := os.Stdout

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}

	os.Stdout = w

	fn()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}

	if err := r.Close(); err != nil {
		t.Fatal(err)
	}

	return buf.String()
}

func TestShowCmdHandleHelp(t *T) {
	node := createTestNode(t)

	cmd := CLIShow(node)

	output := captureOutput(t, func() {
		cmd.handle([]string{"show"})
	})

	expected := cmd.getHelp()

	if !strings.Contains(output, expected) {
		t.Fatalf(
			"expected output to contain %q, got %q",
			expected,
			output,
		)
	}
}

func TestShowCmdHandleRoutingTable(t *T) {
	simulation := network.NewSimulation(0, 0, 1)

	config := kademlia.KademliaConfig{
		Alpha:   3,
		K:       10,
		Timeout: 5 * time.Second,
		Retries: 5,
	}

	mock := &kademlia.Mock{
		Simulation: simulation,
		Config:     config,
	}

	bootNode := mock.CreateNode(0, simulation, config)

	for i := 1; i < 3; i++ {
		node := mock.CreateNode(i, simulation, config)

		go node.Join(bootNode.Me)
	}

	// This is the same synchronization approach your existing mock uses.
	time.Sleep(1 * time.Second)

	cmd := CLIShow(bootNode)

	output := captureOutput(t, func() {
		cmd.handle([]string{"show", "rt"})
	})

	if !strings.Contains(output, "Routing Table:") {
		t.Fatalf("expected routing table header, got:\n%s", output)
	}

	// At least one of the simulated nodes should be present.
	foundNode := false

	for _, node := range mock.Nodes[1:] {
		if strings.Contains(output, node.Me.Address) {
			foundNode = true
			break
		}
	}

	if !foundNode {
		t.Fatalf(
			"expected routing table to contain one of the simulated nodes, got:\n%s",
			output,
		)
	}
}

func TestShowCmdHandleDataStore(t *T) {
	node := createTestNode(t)

	data := []byte("hello world")

	node.Store(data)

	id := contact.NewKademliaIDFromData(data)

	cmd := CLIShow(node)

	output := captureOutput(t, func() {
		cmd.handle([]string{"show", "ds"})
	})

	if !strings.Contains(output, "Keys stored in local data store") {
		t.Fatalf(
			"expected datastore header, got:\n%s",
			output,
		)
	}

	expectedID := truncateId(*id)

	if !strings.Contains(output, expectedID) {
		t.Fatalf(
			"expected datastore output to contain %q, got:\n%s",
			expectedID,
			output,
		)
	}
}

func createTestNode(t *T) *kademlia.Kademlia {
	t.Helper()

	simulation := network.NewSimulation(0, 0, 1)

	config := kademlia.KademliaConfig{
		Alpha:   3,
		K:       10,
		Timeout: 5 * time.Second,
		Retries: 5,
	}

	mock := &kademlia.Mock{
		Simulation: simulation,
		Config:     config,
	}

	return mock.CreateNode(0, simulation, config)
}

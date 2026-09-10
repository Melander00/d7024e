package kademlia

import "testing"

func TestNewRandomKademliaID(t *testing.T) {
	id := NewRandomKademliaID()
	if id == nil {
		t.Fatal("expected a generated Kademlia ID")
	}

	if id.String() == "0000000000000000000000000000000000000000000000000000000000000000" {
		t.Fatal("expected generated Kademlia ID to contain random data")
	}

	if len(id.String()) != IDLength*2 {
		t.Fatalf("expected ID string length %d, got %d", IDLength*2, len(id.String()))
	}
}

package contact

import (
	"testing"
)

type T = testing.T

func TestNewRoutingTable(t *T) {
	me := NewContact(
		parseTestKademliaID(t, "FFFFFFFF00000000000000000000000000000000000000000000000000000000"),
		"localhost:8000",
	)

	k := 10

	rt := NewRoutingTable(me, k)

	if rt == nil {
		t.Fatal("routing table nil")
	}

	for i, bucket := range rt.buckets {
		if bucket == nil {
			t.Fatalf("bucket %d is nil", i)
		}
		if bucket.size != k {
			t.Fatalf("bucket %d has size %d instead of %d", i, bucket.size, k)
		}
	}
}

func TestRoutingTableFindClosestContacts(t *T) {
	me := NewContact(
		parseTestKademliaID(t, "FFFFFFFF00000000000000000000000000000000000000000000000000000000"),
		"localhost:8000",
	)

	rt := NewRoutingTable(me, 10)

	ids := []*KademliaID{
		parseTestKademliaID(t, "1111111100000000000000000000000000000000000000000000000000000000"),
		parseTestKademliaID(t, "1111111200000000000000000000000000000000000000000000000000000000"),
		parseTestKademliaID(t, "1111111300000000000000000000000000000000000000000000000000000000"),
		parseTestKademliaID(t, "1111111400000000000000000000000000000000000000000000000000000000"),
		parseTestKademliaID(t, "2111111400000000000000000000000000000000000000000000000000000000"),
	}

	contacts := []Contact{
		NewContact(
			ids[0],
			"localhost:8001",
		),
		NewContact(
			ids[1],
			"localhost:8002",
		),
		NewContact(
			ids[2],
			"localhost:8003",
		),
		NewContact(
			ids[3],
			"localhost:8004",
		),
		NewContact(
			ids[4],
			"localhost:8005",
		),
	}

	for _, contact := range contacts {
		rt.AddContact(contact)
	}

	target := ids[4]

	closest := rt.FindClosestContacts(target, 3)

	if len(closest) != 3 {
		t.Fatalf("expected 3 contacts, got %d", len(closest))
	}

	// The target itself is included because it is a normal contact.
	expected := []string{
		"2111111400000000000000000000000000000000000000000000000000000000",
		"1111111400000000000000000000000000000000000000000000000000000000",
		"1111111100000000000000000000000000000000000000000000000000000000",
	}

	for i, contact := range closest {
		if !contact.ID.Equals(parseTestKademliaID(t, expected[i])) {
			t.Fatalf(
				"position %d: expected %s, got %s",
				i,
				expected[i],
				contact.ID,
			)
		}
	}
}

func TestRoutingTableDoesNotAddSelf(t *T) {
	me := NewContact(
		parseTestKademliaID(t, "FFFFFFFF00000000000000000000000000000000000000000000000000000000"),
		"localhost:8000",
	)

	rt := NewRoutingTable(me, 10)

	rt.AddContact(me)

	contacts := rt.FindClosestContacts(me.ID, 20)

	if len(contacts) != 0 {
		t.Fatalf("expected 0 contacts, got %d", len(contacts))
	}
}

// FIXME: This test doesn't actually test anything. There is only one assertion
// that is included as an example.

// func TestRoutingTable(t *testing.T) {
// 	rt := NewRoutingTable(NewContact(parseTestKademliaID(t, "FFFFFFFF00000000000000000000000000000000000000000000000000000000"), "localhost:8000"), 10)

// 	rt.AddContact(NewContact(parseTestKademliaID(t, "FFFFFFFF00000000000000000000000000000000000000000000000000000000"), "localhost:8001"))
// 	rt.AddContact(NewContact(parseTestKademliaID(t, "1111111100000000000000000000000000000000000000000000000000000000"), "localhost:8002"))
// 	rt.AddContact(NewContact(parseTestKademliaID(t, "1111111200000000000000000000000000000000000000000000000000000000"), "localhost:8002"))
// 	rt.AddContact(NewContact(parseTestKademliaID(t, "1111111300000000000000000000000000000000000000000000000000000000"), "localhost:8002"))
// 	rt.AddContact(NewContact(parseTestKademliaID(t, "1111111400000000000000000000000000000000000000000000000000000000"), "localhost:8002"))
// 	rt.AddContact(NewContact(parseTestKademliaID(t, "2111111400000000000000000000000000000000000000000000000000000000"), "localhost:8002"))
// 	rt.AddContact(NewContact(parseTestKademliaID(t, "FFFFFFFFFF000000000000000000000000000000000000000000000000000000"), "localhost:8002"))

// 	// Test dont add self to routing table
// 	contacts := rt.FindClosestContacts(parseTestKademliaID(t, "2111111400000000000000000000000000000000000000000000000000000000"), 20)
// 	if len(contacts) != 5 {
// 		t.Fatalf("Expected 5 contacts but instead got %d", len(contacts))
// 	}
// }

func parseTestKademliaID(t *testing.T, value string) *KademliaID {
	t.Helper()

	id, err := ParseKademliaID(value)
	if err != nil {
		t.Fatalf("failed to parse test Kademlia ID: %v", err)
	}

	return id
}

package contact

func TestNewContact(t *T) {
	id, _ := ParseKademliaID("FFFFFFFF00000000000000000000000000000000000000000000000000000000")
	me := NewContact(id, "127.0.0.1:10001")

	if me.ID == nil || me.Address == "" {
		t.Fatalf("ID or address not set %s", me.String())
	}
}

func TestSortNearContactCandidates(t *T) {
	target, _ := ParseKademliaID("0000000000000000000000000000000000000000000000000000000000000000")
	id1, _ := ParseKademliaID("0100000000000000000000000000000000000000000000000000000000000000")
	id2, _ := ParseKademliaID("8000000000000000000000000000000000000000000000000000000000000000")
	id3, _ := ParseKademliaID("4000000000000000000000000000000000000000000000000000000000000000")

	candidates := ContactCandidates{}
	candidates.Append([]Contact{
		// Unordered
		NewContact(id2, "127.0.0.1:10002"),
		NewContact(id1, "127.0.0.1:10001"),
		NewContact(id3, "127.0.0.1:10003"),
	})

	candidates.SortNear(target)

	expected := []*KademliaID{
		id1,
		id3,
		id2,
	}

	for i, c := range candidates.GetContacts(10) {
		if !c.ID.Equals(expected[i]) {
			t.Fatalf("contact %d: expected %s, got %s", i, expected[i], c.ID)
		}
	}
}

func TestRemoveMeContactCandidates(t *T) {
	id, _ := ParseKademliaID("FFFFFFFF00000000000000000000000000000000000000000000000000000000")
	me := NewContact(id, "127.0.0.1:10001")

	candidates := ContactCandidates{}
	candidates.Append([]Contact{
		me,
	})
	candidates.RemoveMe(&me)

	for _, c := range candidates.GetContacts(10) {
		if !c.ID.Equals(me.ID) {
			t.Fatalf("owner was not removed")
		}
	}
}

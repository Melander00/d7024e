package contact

import (
	"fmt"
	"sort"
)

// Contact definition
// stores the KademliaID, the ip address and the distance
type Contact struct {
	ID       *KademliaID
	Address  string
	distance *KademliaID
}

// NewContact returns a new instance of a Contact
func NewContact(id *KademliaID, address string) Contact {
	return Contact{id, address, nil}
}

// CalcDistance calculates the distance to the target and
// fills the contacts distance field
func (contact *Contact) CalcDistance(target *KademliaID) {
	contact.distance = contact.ID.CalcDistance(target)
}

// Less returns true if contact.distance < otherContact.distance
func (contact *Contact) Less(otherContact *Contact) bool {
	return contact.distance.Less(otherContact.distance)
}

// String returns a simple string representation of a Contact
func (contact *Contact) String() string {
	return fmt.Sprintf(`contact("%s", "%s")`, contact.ID, contact.Address)
}

// ContactCandidates definition
// stores an array of Contacts
type ContactCandidates struct {
	contacts []Contact
}

// Append an array of Contacts to the ContactCandidates
//
//	func (candidates *ContactCandidates) Append(contacts []Contact) {
//		// TODO: Deduplicating so that we dont store the same contact multiple times.
//		// Use ID to remove duplicates.
//		candidates.contacts = append(candidates.contacts, contacts...)
//	}

// Append an array of Contacts to the ContactCandidates with deduplicating.
func (candidates *ContactCandidates) Append(contacts []Contact) {
	seen := make(map[KademliaID]struct{}, len(candidates.contacts)+len(contacts))

	for _, contact := range candidates.contacts {
		if contact.ID != nil {
			seen[*contact.ID] = struct{}{}
		}
	}

	for _, contact := range contacts {
		if contact.ID == nil {
			continue
		}

		if _, exists := seen[*contact.ID]; exists {
			continue
		}

		seen[*contact.ID] = struct{}{}
		candidates.contacts = append(candidates.contacts, contact)
	}
}

// GetContacts returns the first count number of Contacts
func (candidates *ContactCandidates) GetContacts(count int) []Contact {
	m := min(candidates.Len(), count)
	return candidates.contacts[:m]
}

// Sort the Contacts in ContactCandidates
func (candidates *ContactCandidates) Sort() {
	sort.Sort(candidates)
}

func (candidates *ContactCandidates) SortNear(target *KademliaID) {
	for i := range candidates.contacts {
		candidates.contacts[i].CalcDistance(target)
	}

	sort.Sort(candidates)
}

// Len returns the length of the ContactCandidates
func (candidates *ContactCandidates) Len() int {
	return len(candidates.contacts)
}

// Swap the position of the Contacts at i and j
// WARNING does not check if either i or j is within range
func (candidates *ContactCandidates) Swap(i, j int) {
	candidates.contacts[i], candidates.contacts[j] = candidates.contacts[j], candidates.contacts[i]
}

// Less returns true if the Contact at index i is smaller than
// the Contact at index j
func (candidates *ContactCandidates) Less(i, j int) bool {
	return candidates.contacts[i].Less(&candidates.contacts[j])
}

func (candidates *ContactCandidates) RemoveMe(me *Contact) {
	for i, cand := range candidates.contacts {
		if cand.ID.Equals(me.ID) {
			candidates.contacts = append(candidates.contacts[:i], candidates.contacts[i+1:]...)

			return
		}
	}
}

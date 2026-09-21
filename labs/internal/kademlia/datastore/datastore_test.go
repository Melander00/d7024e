package datastore

import (
	"bytes"
	"d7024e/internal/kademlia/contact"
	"testing"
)

// TestDataStorePutAndGet verifies that a value can be stored in the datastore
// using a Kademlia ID as its key and then retrieved using the same key.
// The test checks both that the key is reported as existing and that the
// retrieved value is identical to the value that was originally stored.

func TestDataStorePutAndGet(t *testing.T) {
	ds := NewDataStore()

	key := contact.NewRandomKademliaID()
	expected := []byte("hello")

	err := ds.Put(key, expected)
	if err != nil {
		t.Fatal(err)
	}

	actual, exists := ds.Get(key)
	if !exists {
		t.Fatal("expected key to exist")
	}

	if !bytes.Equal(actual, expected) {
		t.Fatalf("expected %q, got %q", expected, actual)
	}

}

func TestDataStoreGetWithEquivalentKey(t *testing.T) {
	ds := NewDataStore()

	storedKey := contact.NewKademliaIDFromData([]byte("hello"))
	lookupKey, err := contact.ParseKademliaID(storedKey.String())
	if err != nil {
		t.Fatal(err)
	}

	if storedKey == lookupKey {
		t.Fatal("expected stored and lookup keys to be distinct pointers")
	}

	expected := []byte("hello")
	if err := ds.Put(storedKey, expected); err != nil {
		t.Fatal(err)
	}

	actual, exists := ds.Get(lookupKey)
	if !exists {
		t.Fatal("expected equivalent key to retrieve stored value")
	}

	if !bytes.Equal(actual, expected) {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}

// TestDataStoreGetMissingKey verifies that requesting a key that has not been stored
// reports that the key does not exist.
func TestDataStoreGetMissingKey(t *testing.T) {
	ds := NewDataStore()
	key := contact.NewRandomKademliaID()
	actual, exists := ds.Get(key)
	if exists {
		t.Fatal("expected key to not exist")
	}
	if actual != nil {
		t.Fatalf("expected nil value for missing key, got %v", actual)
	}
}

// TestDataStoreKeys verifies that Keys returns the keys currently stored in
// the datastore after multiple values have been inserted.
func TestDataStoreKeys(t *testing.T) {
	ds := NewDataStore()

	key1 := contact.NewRandomKademliaID()
	key2 := contact.NewRandomKademliaID()

	if err := ds.Put(key1, []byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := ds.Put(key2, []byte("world")); err != nil {
		t.Fatal(err)
	}

	keys := ds.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}

	foundKey1 := false
	foundKey2 := false
	for _, key := range keys {
		if key.Equals(key1) {
			foundKey1 = true
		}
		if key.Equals(key2) {
			foundKey2 = true
		}
	}

	if !foundKey1 || !foundKey2 {
		t.Fatal("expected Keys to return both stored keys")
	}

}

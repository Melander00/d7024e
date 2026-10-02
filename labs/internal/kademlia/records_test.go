package kademlia

import (
	"bytes"
	"d7024e/internal/kademlia/contact"
	"d7024e/internal/kademlia/datastore"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"
)

func TestVersionRecordSerializationAndStorage(t *testing.T) {
	record := NewVersionRecord("example.com", "package", 2, "blob-hash", "previous-hash", []byte{1, 2})
	expected := `{"tag":"version-record","domain-name":"example.com","package-name":"package","version":2,"blobHash":"blob-hash","previous-version-record":"previous-hash","sig":"AQI="}`

	if got := string(record.Serialize()); got != expected {
		t.Fatalf("unexpected serialization: got %q, want %q", got, expected)
	}
	if got := string(record.SigningBytes()); got != `{"tag":"version-record","domain-name":"example.com","package-name":"package","version":2,"blobHash":"blob-hash","previous-version-record":"previous-hash"}` {
		t.Fatalf("unexpected signing bytes: %q", got)
	}

	store := datastore.NewDataStore()
	if err := record.Store(store); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadVersionRecord(store, record.Hash())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(record, loaded) {
		t.Fatalf("record changed after round trip: got %#v, want %#v", loaded, record)
	}

	expectedHash := contact.NewKademliaIDFromData([]byte(expected))
	if !record.Hash().Equals(expectedHash) {
		t.Fatalf("unexpected record hash: got %s, want %s", record.Hash().String(), expectedHash.String())
	}
}

func TestLatestPointerUsesStableMutableKey(t *testing.T) {
	pointer := NewLatestPointer("example.com", "package", 2, "record-hash", []byte{3})
	store := datastore.NewDataStore()
	if err := pointer.Store(store); err != nil {
		t.Fatal(err)
	}

	loaded, err := LoadLatestPointer(store, pointer.Key())
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(pointer, loaded) {
		t.Fatalf("pointer changed after round trip: got %#v, want %#v", loaded, pointer)
	}

	newPointer := NewLatestPointer("example.com", "package", 3, "new-record-hash", []byte{4})
	if !pointer.Key().Equals(newPointer.Key()) {
		t.Fatal("latest pointer key changed with the pointed-to version")
	}
	if pointer.Hash().Equals(newPointer.Hash()) {
		t.Fatal("different pointer values unexpectedly have the same content hash")
	}
}

func TestRecordHashesAreIndependentOfDecodedValues(t *testing.T) {
	record := NewVersionRecord("example.com", "package", 1, "blob", "zero", []byte{9})
	encoded := record.Serialize()
	var decoded VersionRecord
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, decoded.Serialize()) {
		t.Fatalf("serialization changed after decode: %q != %q", encoded, decoded.Serialize())
	}
	if got := hex.EncodeToString(record.Hash()[:]); got == "" {
		t.Fatal("record hash was empty")
	}
}

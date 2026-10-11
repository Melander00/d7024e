package kademlia

import (
	"d7024e/internal/kademlia/contact"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"
)

const (
	part2TestDomain  = "part2.example.test"
	part2TestPackage = "demo"
)

// Creates one simulated node and registers its signing key as the
// owner of the test domain.
func newPart2Node(t *testing.T) *Kademlia {
	t.Helper()

	mock := MockKademlia(t, 1)
	node := mock.Nodes[0]

	if err := node.DNS.ClaimDomain(
		part2TestDomain,
		node.Signature.GetPublicKey(),
	); err != nil {
		t.Fatalf("ClaimDomain: %v", err)
	}

	return node
}

// Creates a signed record and stores it under its content hash.
// previousHash must be zeroHash() for the genesis record.
func addPart2Version(
	t *testing.T,
	node *Kademlia,
	domain string,
	pkg string,
	version uint64,
	previousHash string,
) *VersionRecord {
	t.Helper()

	blob := []byte(fmt.Sprintf("package-%s-version-%d", pkg, version))
	blobHash := contact.NewKademliaIDFromData(blob)

	if err := node.Datastore.Put(blobHash, blob); err != nil {
		t.Fatalf("store blob: %v", err)
	}

	record := NewVersionRecord(
		domain,
		pkg,
		version,
		blobHash.String(),
		previousHash,
		nil,
	)
	record = node.Signature.SignVersionRecord(record)

	if err := node.Datastore.Put(record.Hash(), record.Serialize()); err != nil {
		t.Fatalf("store version record: %v", err)
	}

	return &record
}

func makePart2Pointer(
	node *Kademlia,
	domain string,
	pkg string,
	version uint64,
	recordHash string,
) LatestPointer {
	pointer := NewLatestPointer(
		domain,
		pkg,
		version,
		recordHash,
		nil,
	)

	return node.Signature.SignLatestRecord(pointer)
}

// Stores a proposed latest pointer through the same handler used for
// incoming STORE requests.
func submitPart2Pointer(
	node *Kademlia,
	pointer LatestPointer,
) bool {
	return node.storeHandler(
		node.Me,
		pointer.Key().String(),
		pointer.Serialize(),
	)
}

func TestPart2StoreAcceptsValidGenesisPointer(t *testing.T) {
	node := newPart2Node(t)

	genesis := addPart2Version(
		t,
		node,
		part2TestDomain,
		part2TestPackage,
		1,
		zeroHash(),
	)

	pointer := makePart2Pointer(
		node,
		part2TestDomain,
		part2TestPackage,
		1,
		genesis.Hash().String(),
	)

	if !submitPart2Pointer(node, pointer) {
		t.Fatal("valid signed genesis pointer was rejected")
	}

	data, exists := node.Datastore.Get(pointer.Key())
	if !exists {
		t.Fatal("latest pointer was not stored")
	}
	if len(data) == 0 {
		t.Fatal("stored latest pointer is empty")
	}
}

func TestPart2StoreRejectsTamperedPointer(t *testing.T) {
	node := newPart2Node(t)

	genesis := addPart2Version(
		t,
		node,
		part2TestDomain,
		part2TestPackage,
		1,
		zeroHash(),
	)

	pointer := makePart2Pointer(
		node,
		part2TestDomain,
		part2TestPackage,
		1,
		genesis.Hash().String(),
	)

	// Change a signed field after signing.
	pointer.Version = 2

	if submitPart2Pointer(node, pointer) {
		t.Fatal("tampered pointer signature was accepted")
	}
}

func TestPart2StoreRejectsRollback(t *testing.T) {
	node := newPart2Node(t)

	v1 := addPart2Version(
		t, node, part2TestDomain, part2TestPackage, 1, zeroHash(),
	)
	v2 := addPart2Version(
		t, node, part2TestDomain, part2TestPackage, 2, v1.Hash().String(),
	)

	p2 := makePart2Pointer(
		node, part2TestDomain, part2TestPackage, 2, v2.Hash().String(),
	)
	if !submitPart2Pointer(node, p2) {
		t.Fatal("valid version 2 pointer was rejected")
	}

	p1 := makePart2Pointer(
		node, part2TestDomain, part2TestPackage, 1, v1.Hash().String(),
	)
	if submitPart2Pointer(node, p1) {
		t.Fatal("rollback from version 2 to version 1 was accepted")
	}
}

func TestPart2StoreRejectsFork(t *testing.T) {
	node := newPart2Node(t)

	v1 := addPart2Version(
		t, node, part2TestDomain, part2TestPackage, 1, zeroHash(),
	)
	v2 := addPart2Version(
		t, node, part2TestDomain, part2TestPackage, 2, v1.Hash().String(),
	)

	current := makePart2Pointer(
		node, part2TestDomain, part2TestPackage, 2, v2.Hash().String(),
	)
	if !submitPart2Pointer(node, current) {
		t.Fatal("valid version 2 pointer was rejected")
	}

	// This branch skips H2 and points directly back to H1.
	fork := addPart2Version(
		t, node, part2TestDomain, part2TestPackage, 3, v1.Hash().String(),
	)
	proposed := makePart2Pointer(
		node, part2TestDomain, part2TestPackage, 3, fork.Hash().String(),
	)

	if submitPart2Pointer(node, proposed) {
		t.Fatal("fork that does not extend the current head was accepted")
	}
}

func TestPart2StoreAcceptsCatchUp(t *testing.T) {
	node := newPart2Node(t)

	v1 := addPart2Version(
		t, node, part2TestDomain, part2TestPackage, 1, zeroHash(),
	)
	v2 := addPart2Version(
		t, node, part2TestDomain, part2TestPackage, 2, v1.Hash().String(),
	)
	v3 := addPart2Version(
		t, node, part2TestDomain, part2TestPackage, 3, v2.Hash().String(),
	)

	// Start with a locally stored version 1 pointer.
	p1 := makePart2Pointer(
		node, part2TestDomain, part2TestPackage, 1, v1.Hash().String(),
	)
	if !submitPart2Pointer(node, p1) {
		t.Fatal("valid version 1 pointer was rejected")
	}

	// The node missed version 2 and receives a valid version 3 chain.
	p3 := makePart2Pointer(
		node, part2TestDomain, part2TestPackage, 3, v3.Hash().String(),
	)
	if !submitPart2Pointer(node, p3) {
		t.Fatal("valid catch-up update was rejected")
	}

	data, exists := node.Datastore.Get(p3.Key())
	if !exists {
		t.Fatal("latest pointer missing after catch-up")
	}

	var stored LatestPointer
	if err := unmarshalPart2Pointer(data, &stored); err != nil {
		t.Fatalf("decode stored pointer: %v", err)
	}
	if stored.Version != 3 || stored.VersionRecordHash != v3.Hash().String() {
		t.Fatal("stored pointer does not reference version 3")
	}
}

func TestPart2StoreRejectsMissingVersionRecord(t *testing.T) {
	node := newPart2Node(t)

	// The hash is syntactically valid, but no record is stored at this key.
	missingHash := contact.NewRandomKademliaID().String()

	pointer := makePart2Pointer(
		node,
		part2TestDomain,
		part2TestPackage,
		1,
		missingHash,
	)

	if submitPart2Pointer(node, pointer) {
		t.Fatal("pointer referencing a missing version record was accepted")
	}
}

func TestPart2GetLatestVersionRejectsWrongPackageIdentity(t *testing.T) {
	node := newPart2Node(t)

	genesis := addPart2Version(
		t, node, part2TestDomain, part2TestPackage, 1, zeroHash(),
	)
	pointer := makePart2Pointer(
		node, part2TestDomain, part2TestPackage, 1, genesis.Hash().String(),
	)

	// Store under the correct deterministic key, but query a different
	// package. The pointer must not be treated as that package's latest.
	if !submitPart2Pointer(node, pointer) {
		t.Fatal("could not store valid test pointer")
	}

	_, err := node.GetLatestVersion(part2TestDomain, "different-package")
	if err == nil {
		t.Fatal("expected lookup of a different package to fail")
	}
}

func TestPart2GetVersionRecordRejectsUnknownHash(t *testing.T) {
	node := newPart2Node(t)

	_, err := node.GetVersionRecord(
		contact.NewRandomKademliaID().String(),
	)
	if err == nil {
		t.Fatal("expected lookup of a missing version record to fail")
	}
}

func TestPart2VersionRecordHistoryRejectsMissingPredecessor(t *testing.T) {
	node := newPart2Node(t)

	missingPrevious := contact.NewRandomKademliaID().String()
	record := addPart2Version(
		t,
		node,
		part2TestDomain,
		part2TestPackage,
		2,
		missingPrevious,
	)

	pointer := makePart2Pointer(
		node,
		part2TestDomain,
		part2TestPackage,
		2,
		record.Hash().String(),
	)

	if submitPart2Pointer(node, pointer) {
		t.Fatal("pointer with a missing predecessor was accepted")
	}
}

func TestPart2StoreAcceptsIdenticalPointerIdempotently(t *testing.T) {
	node := newPart2Node(t)

	genesis := addPart2Version(
		t, node, part2TestDomain, part2TestPackage, 1, zeroHash(),
	)
	pointer := makePart2Pointer(
		node, part2TestDomain, part2TestPackage, 1, genesis.Hash().String(),
	)

	if !submitPart2Pointer(node, pointer) {
		t.Fatal("first submission of valid pointer was rejected")
	}
	if !submitPart2Pointer(node, pointer) {
		t.Fatal("identical repeat pointer should be idempotently accepted")
	}
}

// This wrapper keeps the test file independent of any JSON helper used
// elsewhere in the project.
func unmarshalPart2Pointer(data []byte, pointer *LatestPointer) error {
	if len(data) == 0 {
		return errors.New("empty pointer data")
	}

	return jsonUnmarshal(data, pointer)
}

// Assigned below through encoding/json's standard decoder.
var jsonUnmarshal = func(data []byte, pointer *LatestPointer) error {
	return json.Unmarshal(data, pointer)
}

// Ensure the mock configuration remains suitable for these tests.
func TestPart2MockConfig(t *testing.T) {
	config := KademliaConfig{
		Alpha:   3,
		K:       10,
		Timeout: 5 * time.Second,
		Retries: 1,
	}
	if config.Alpha <= 0 || config.K <= 0 || config.Timeout <= 0 {
		t.Fatal("invalid test configuration")
	}
}
func TestPart2StoreAcceptsCatchUpFromAnotherNode(t *testing.T) {
	m := MockKademlia(t, 2)

	publisher := m.Nodes[0]
	receiver := m.Nodes[1]

	publisher.DNS.ClaimDomain(part2TestDomain, publisher.Signature.GetPublicKey())

	// Create the complete history on the publisher only.

	println("=== 1 ===")

	v1 := addPart2Version(
		t, publisher, part2TestDomain, part2TestPackage, 1, zeroHash(),
	)
	v2 := addPart2Version(
		t, publisher, part2TestDomain, part2TestPackage, 2, v1.Hash().String(),
	)
	v3 := addPart2Version(
		t, publisher, part2TestDomain, part2TestPackage, 3, v2.Hash().String(),
	)

	fmt.Printf("v1 %s\n", v1.Hash().String())
	fmt.Printf("v2 %s\n", v2.Hash().String())
	fmt.Printf("v3 %s\n", v3.Hash().String())

	println("=== 2 ===")
	// Configure routing so the nodes can discover each other.
	publisher.Routing.AddContact(receiver.Me)
	receiver.Routing.AddContact(publisher.Me)

	// receiver.PrintRT()

	// The receiver starts with only version 1's record locally.
	// Version 2 and version 3 remain on the publisher.
	if err := receiver.Datastore.Put(v1.Hash(), v1.Serialize()); err != nil {
		t.Fatalf("store version 1 on receiver: %v", err)
	}

	println("=== 3 ===")

	for _, hash := range []string{
		v2.Hash().String(),
		v3.Hash().String(),
	} {
		key, err := contact.ParseKademliaID(hash)
		if err != nil {
			t.Fatalf("parse version hash %q: %v", hash, err)
		}
		if _, exists := receiver.Datastore.Get(key); exists {
			t.Fatalf("receiver unexpectedly has version record %s", hash)
		}
	}

	println("=== 5 ===")

	// Install version 1 as the receiver's current pointer.
	p1 := makePart2Pointer(
		publisher,
		part2TestDomain,
		part2TestPackage,
		1,
		v1.Hash().String(),
	)

	println("=== 6 ===")

	if !receiver.storeHandler(
		publisher.Me,
		p1.Key().String(),
		p1.Serialize(),
	) {
		t.Fatal("receiver rejected valid version 1 pointer")
	}

	println("=== 7 ===")

	// The receiver missed the updates to versions 2 and 3.
	// Submitting version 3 must cause history validation to fetch the
	// missing version records from the publisher through the DHT.
	p3 := makePart2Pointer(
		publisher,
		part2TestDomain,
		part2TestPackage,
		3,
		v3.Hash().String(),
	)

	println("=== 8 ===")

	if !receiver.storeHandler(
		publisher.Me,
		p3.Key().String(),
		p3.Serialize(),
	) {
		t.Fatal("receiver rejected valid version 3 catch-up pointer")
	}

	println("=== 9 ===")

	// Verify that the receiver advanced its latest pointer to version 3.
	data, exists := receiver.Datastore.Get(p3.Key())
	if !exists {
		t.Fatal("receiver did not store the version 3 latest pointer")
	}

	println("=== 10 ===")

	var stored LatestPointer
	if err := unmarshalPart2Pointer(data, &stored); err != nil {
		t.Fatalf("decode stored pointer: %v", err)
	}

	if stored.Version != 3 {
		t.Fatalf("stored version = %d, want 3", stored.Version)
	}
	if stored.VersionRecordHash != v3.Hash().String() {
		t.Fatalf(
			"stored record hash = %s, want %s",
			stored.VersionRecordHash,
			v3.Hash().String(),
		)
	}

	println("=== 11 ===")
}

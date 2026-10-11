package kademlia

import (
	"d7024e/internal/kademlia/contact"
	"d7024e/internal/kademlia/datastore"
	"d7024e/internal/kademlia/rpc"
	d "d7024e/pkg/dns"
	"d7024e/pkg/logger"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"
)

type Kademlia struct {
	Rpc       *rpc.RPC
	ID        *contact.KademliaID
	Routing   *contact.RoutingTable
	Me        contact.Contact
	Datastore *datastore.DataStore
	Config    *KademliaConfig
	Logger    logger.Logger
	Signature Signature
	DNS       d.DNS
}

type KademliaConfig struct {
	Alpha                 int
	K                     int
	Timeout               time.Duration
	Retries               int
	BucketRefreshInterval time.Duration
	ReplicationInterval   time.Duration
}

// defaultBucketRefreshInterval is used when a config leaves BucketRefreshInterval unset. set to half an hour to guarantee periodic refresh of buckets of an hour
const defaultBucketRefreshInterval = time.Hour

// defaultReplicationInterval is used when a config leaves ReplicationInterval unset.
const defaultReplicationInterval = time.Hour

func NewKademliaNode(config KademliaConfig, rpc *rpc.RPC) *Kademlia {

	id, _ := contact.NewKademliaIDFromAddress(rpc.Me.Address)

	rpc.SetTimeout(config.Timeout)
	rpc.SetRetries(config.Retries)

	if config.BucketRefreshInterval <= 0 {
		config.BucketRefreshInterval = defaultBucketRefreshInterval
	}
	if config.ReplicationInterval <= 0 {
		config.ReplicationInterval = defaultReplicationInterval
	}

	routing := contact.NewRoutingTable(rpc.Me, config.K)

	sign, _ := NewEd25519Sign()

	kademlia := &Kademlia{
		Rpc:       rpc,
		ID:        id,
		Me:        rpc.Me,
		Routing:   routing,
		Datastore: datastore.NewDataStore(),
		Config:    &config,
		Logger:    logger.NewEmptyLogger(),
		Signature: sign,
		DNS:       d.NewRealDNS(os.Getenv("DNS")),
	}

	rpc.SetPingHandler(kademlia.pingHandler)
	rpc.SetFindNodeHandler(kademlia.findNodeHandler)
	rpc.SetFindValueHandler(kademlia.findValueHandler)
	rpc.SetStoreHandler(kademlia.storeHandler)

	return kademlia
}

type lookupResult struct {
	contacts []contact.Contact
	value    []byte
	found    bool
	err      error
	contact  contact.Contact
}

type lookupQuery func(contact.Contact, *contact.KademliaID) lookupResult

func (kademlia *Kademlia) lookup(target *contact.KademliaID, query lookupQuery) ([]contact.Contact, []byte, error) {
	candidates := contact.ContactCandidates{}
	candidates.Append(kademlia.Routing.FindClosestContacts(target, kademlia.Config.K))

	candidates.RemoveMe(&kademlia.Me)

	candidates.Sort()

	queried := make(map[string]bool)
	results := make(chan lookupResult, kademlia.Config.Alpha)

	pending := 0

	for {

		for pending < kademlia.Config.Alpha {

			candidate, ok := getNextCandidate(&candidates, queried, kademlia.Config.K)

			if !ok {
				// We have reached the end
				break
			}

			queried[candidate.ID.String()] = true
			pending++

			go func(candidate contact.Contact) {
				results <- query(candidate, target)
			}(candidate)
		}

		if pending == 0 {
			// If we have no more contacts to request and no more inflight requests we stop
			break
		}

		result := <-results
		pending--

		if result.err != nil {
			// TODO: error handling
			// If the error is timeout we can assume it is dead for example
			kademlia.Routing.RemoveContact(result.contact)
			continue
		}

		for _, c := range result.contacts {
			kademlia.Routing.AddContact(c)
		}
		kademlia.Routing.AddContact(result.contact)

		if result.found {
			return nil, result.value, nil
		}

		candidates.Append(result.contacts)
		candidates.SortNear(target)
	}

	// Return the k closest known contacts.
	if candidates.Len() > kademlia.Config.K {
		return candidates.GetContacts(kademlia.Config.K), nil, nil
	}

	return candidates.GetContacts(candidates.Len()), nil, nil
}

func getNextCandidate(candidates *contact.ContactCandidates, queried map[string]bool, k int) (contact.Contact, bool) {
	contacts := candidates.GetContacts(candidates.Len())

	for i, candidate := range contacts {
		if !queried[candidate.ID.String()] {
			return candidate, true
		}
		if i >= k {
			break
		}
	}

	return contact.Contact{}, false
}

func (kademlia *Kademlia) LookupContact(target *contact.Contact) []contact.Contact {
	// Self-lookups have no bucket (XOR distance is zero)
	if !target.ID.Equals(kademlia.ID) {
		kademlia.Routing.MarkLookup(target.ID)
	}

	kademlia.Logger.Log(fmt.Sprintf("%s lookup_contact_start %s\n", kademlia.Me.Address, target.ID.String()))
	defer kademlia.Logger.Log(fmt.Sprintf("%s lookup_contact_end\n", kademlia.Me.Address))

	i := 0
	mut := &sync.Mutex{}

	contacts, _, _ := kademlia.lookup(
		target.ID,
		func(candidate contact.Contact, target *contact.KademliaID) lookupResult {
			mut.Lock()
			i++
			reqI := i
			mut.Unlock()
			kademlia.Logger.Log(fmt.Sprintf("%s lookup_contact_rpc %d %s\n", kademlia.Me.Address, reqI, candidate.ID.String()))
			res, err := kademlia.Rpc.FindNode(candidate, target.String())

			if err != nil {
				kademlia.Logger.Log(fmt.Sprintf("%s lookup_contact_rpc_error %d %s\n", kademlia.Me.Address, reqI, err.Error()))
				return lookupResult{
					contact: candidate,
					err:     err,
				}
			}

			return lookupResult{
				contact:  candidate,
				contacts: res.Value.Nodes,
			}
		},
	)

	return contacts
}

func (kademlia *Kademlia) LookupData(hash string) ([]byte, error) {
	return kademlia.lookupData(hash, true)
}

func (kademlia *Kademlia) LookupDataAtKey(key string) ([]byte, error) {
	return kademlia.lookupData(key, false)
}

func (kademlia *Kademlia) lookupData(hash string, contentAddressed bool) ([]byte, error) {
	target, err := contact.ParseKademliaID(hash)

	if err != nil {
		return nil, err
	}

	if data, ok := kademlia.Datastore.Get(target); ok {
		return data, nil
	}

	kademlia.Logger.Log(fmt.Sprintf("%s lookup_data_start %s\n", kademlia.Me.Address, hash))
	defer kademlia.Logger.Log(fmt.Sprintf("%s lookup_data_end\n", kademlia.Me.Address))

	i := 0
	mut := &sync.Mutex{}

	_, value, err := kademlia.lookup(
		target,
		func(candidate contact.Contact, target *contact.KademliaID) lookupResult {
			mut.Lock()
			i++
			reqI := i
			mut.Unlock()
			kademlia.Logger.Log(fmt.Sprintf("%s lookup_data_rpc %d %s\n", kademlia.Me.Address, reqI, candidate.ID.String()))
			res, err := kademlia.Rpc.FindValue(candidate, target.String())

			// fmt.Printf("%d 1 %d \n", reqI, len(res.Value.Value))

			if err != nil {
				kademlia.Logger.Log(fmt.Sprintf("%s lookup_data_rpc_error %d %s\n", kademlia.Me.Address, reqI, err.Error()))
				return lookupResult{
					contact: candidate,
					err:     err,
				}
			}

			// fmt.Printf("%d 2 %d \n", reqI, len(res.Value.Value))

			if res.Value.Value != nil {
				// fmt.Printf("%d 2.1 %d \n", reqI, len(res.Value.Value))

				if !contentAddressed || contact.NewKademliaIDFromData(res.Value.Value).Equals(target) {
					// fmt.Printf("%d 2.1.1 %d \n", reqI, len(res.Value.Value))

					kademlia.Logger.Log(fmt.Sprintf("%s lookup_data_rpc_value %d %d\n", kademlia.Me.Address, reqI, len(res.Value.Value)))
					return lookupResult{
						contact: candidate,
						value:   res.Value.Value,
						found:   true,
					}
				}

				// fmt.Printf("%d 2.2 %d \n", reqI, len(res.Value.Value))

				return lookupResult{
					contact: candidate,
					err:     fmt.Errorf("kademlia: value for %s does not hash to the requested key", target.String()),
				}
			}

			// fmt.Printf("%d 3 %d \n", reqI, len(res.Value.Value))

			return lookupResult{
				contact:  candidate,
				contacts: res.Value.Nodes,
			}
		},
	)

	return value, err
}

func (kademlia *Kademlia) Store(data []byte) {
	kademlia.StoreAtKey(contact.NewKademliaIDFromData(data), data)
}

func (kademlia *Kademlia) StoreAtKey(key *contact.KademliaID, data []byte) {

	// fmt.Printf("Store len %d: %s\n", len(data), key.String())

	if err := kademlia.Datastore.Put(key, data); err != nil {
		fmt.Printf("STORE ERROR %s\n", err.Error())
		return
	}

	contacts := kademlia.LookupContact(&contact.Contact{ID: key})

	for _, candidate := range contacts {
		if candidate.Address == kademlia.Me.Address {
			continue
		}
		_, _ = kademlia.Rpc.Store(candidate, key.String(), data)
	}
}

func (kademlia *Kademlia) StoreLatestPointer(pointer LatestPointer) {
	kademlia.StoreAtKey(pointer.Key(), pointer.Serialize())
}

func (kademlia *Kademlia) pingHandler(client contact.Contact) {
	kademlia.Routing.AddContact(client)
	return // Ping shouldnt do anything.
}

func (kademlia *Kademlia) findNodeHandler(
	client contact.Contact,
	hash string,
) ([]contact.Contact, error) {

	target, err := contact.ParseKademliaID(hash)

	if err != nil {
		return nil, err
	}

	kademlia.Routing.AddContact(client)

	contacts :=
		kademlia.Routing.FindClosestContacts(
			target,
			kademlia.Config.K,
		)

	return contacts, nil
}

func (kademlia *Kademlia) findValueHandler(
	client contact.Contact,
	hash string,
) ([]contact.Contact, []byte, bool, error) {

	key, err := contact.ParseKademliaID(hash)

	if err != nil {
		return nil, nil, false, err
	}

	kademlia.Routing.AddContact(client)

	data, has := kademlia.Datastore.Get(key)

	if has {
		return nil, data, true, nil
	}

	contacts :=
		kademlia.Routing.FindClosestContacts(
			key,
			kademlia.Config.K,
		)

	return contacts, nil, false, nil
}

func (kademlia *Kademlia) storeHandler(client contact.Contact, key string, value []byte) bool {
	kademlia.Routing.AddContact(client)
	target, err := contact.ParseKademliaID(key)
	if err != nil {
		return false
	}

	expected := contact.NewKademliaIDFromData(value)
	if target.Equals(expected) {
		return kademlia.Datastore.Put(target, value) == nil
	}

	var pointer LatestPointer
	if err := json.Unmarshal(value, &pointer); err != nil || pointer.Tag != latestPointerTag || !pointer.Key().Equals(target) {
		return false
	}
	if pointer.Version == 0 || len(pointer.Signature) == 0 {
		return false
	}
	if _, err := contact.ParseKademliaID(pointer.VersionRecordHash); err != nil {
		return false
	}

	pk, err := kademlia.DNS.LookupPK(pointer.DomainName)
	if err != nil {
		return false
	}

	if !kademlia.Signature.VerifyLatestRecord(pointer, pk) {
		return false
	}

	// VALIDATE VERSION

	var current *LatestPointer

	if existingBytes, ok := kademlia.Datastore.Get(target); ok {
		var existing LatestPointer
		if err := json.Unmarshal(existingBytes, &existing); err != nil {
			return false
		}
		current = &existing
	}

	if err := kademlia.validatePointerUpdate(pointer, current); err != nil {
		fmt.Println(err)
		return false
	}

	return kademlia.Datastore.Put(target, value) == nil

	// if currentData, exists := kademlia.Datastore.Get(target); exists {
	// 	var current LatestPointer
	// 	if err := json.Unmarshal(currentData, &current); err != nil || current.Tag != latestPointerTag {
	// 		return false
	// 	}
	// 	if pointer.Version <= current.Version {
	// 		return false
	// 	}
	// }

	// return kademlia.Datastore.Put(target, value) == nil
}

func (kademlia *Kademlia) Join(boot contact.Contact) {
	kademlia.Routing.AddContact(boot)
	kademlia.LookupContact(&kademlia.Me)
	// Find closest node we now know.
	closest := kademlia.Routing.FindClosestContacts(
		kademlia.ID,
		1,
	)
	// TODO: return error
	if len(closest) == 0 {
		return
	}

	// Find the bucket containing that closest neighbor.
	closestBucket :=
		kademlia.Routing.BucketIndex(closest[0].ID)

	// Our indexing has farther buckets at smaller indexes.
	for bucketIndex := 0; bucketIndex < closestBucket; bucketIndex++ {

		kademlia.refreshBucket(bucketIndex)
	}

}

func (kademlia *Kademlia) Publish(domain string, pkg string, version uint64, data []byte, force bool, prev string) error {
	prevHash := ""

	pk, err := kademlia.DNS.LookupPK(domain)
	if !pk.Equal(kademlia.Signature.GetPublicKey()) {
		return fmt.Errorf("you dont own that domain")
	}

	p, err := kademlia.GetLatestVersion(domain, pkg)
	if err != nil {
		// Package doesnt exist so we skip validation
		force = true
		prevHash = zeroHash()
	} else {
		prevHash = p.VersionRecordHash
	}

	if !force {
		// TODO validate version
		if p.Version+1 != version {
			return fmt.Errorf("Invalid version")
		}
	}

	if prev != "" && len(prev) == 64 && prev == p.VersionRecordHash {
		prevHash = prev
	}

	kademlia.Store(data)

	blobHash := contact.NewKademliaIDFromData(data)

	rec := NewVersionRecord(domain, pkg, version, blobHash.String(), prevHash, []byte(nil))
	rec = kademlia.Signature.SignVersionRecord(rec)

	point := NewLatestPointer(domain, pkg, version, rec.Hash().String(), []byte(nil))
	point = kademlia.Signature.SignLatestRecord(point)

	kademlia.Store(rec.Serialize())
	kademlia.StoreAtKey(point.Key(), point.Serialize())
	// fmt.Printf("stored %s\n", point.Key().String())

	return nil
}

func (kademlia *Kademlia) GetLatestVersion(domain string, pkg string) (*LatestPointer, error) {
	id := contact.NewKademliaIDFromData([]byte(fmt.Sprintf("%s:%s:latest", domain, pkg)))
	// fmt.Printf("get-latest-version: %s\n", id.String())

	data, err := kademlia.LookupDataAtKey(id.String())
	if err != nil {
		return &LatestPointer{}, err
	}

	if len(data) == 0 {
		// fmt.Printf("data is zero %s\n", data)
		return &LatestPointer{}, fmt.Errorf("package not found")
	}

	pointer := &LatestPointer{}
	err = json.Unmarshal(data, pointer)
	if err != nil {
		return pointer, err
	}

	// verify
	pk, err := kademlia.DNS.LookupPK(domain)
	if err != nil {
		return &LatestPointer{}, err
	}

	// fmt.Printf("data %s\n", pointer)

	ok := kademlia.Signature.VerifyLatestRecord(*pointer, pk)
	if !ok {
		return &LatestPointer{}, errors.New("could not verify signature")
	}

	return pointer, nil
}

func (kademlia *Kademlia) GetVersionRecord(recordHash string) (*VersionRecord, error) {
	data, err := kademlia.LookupData(recordHash)
	if err != nil {
		return &VersionRecord{}, err
	}

	if len(data) == 0 {
		return &VersionRecord{}, errors.New("package not found")
	}

	rec := &VersionRecord{}
	err = json.Unmarshal(data, rec)
	if err != nil {
		return rec, err
	}

	pk, err := kademlia.DNS.LookupPK(rec.DomainName)
	if err != nil {
		return &VersionRecord{}, err
	}

	ok := kademlia.Signature.VerifyVersionRecord(*rec, pk)
	if !ok {
		return &VersionRecord{}, errors.New("could not verify signature")
	}

	return rec, nil
}

func (k *Kademlia) validatePointerUpdate(
	pointer LatestPointer,
	current *LatestPointer,
) error {
	// 1. Validate basic pointer fields.
	if pointer.Tag != latestPointerTag {
		return errors.New("invalid latest pointer tag")
	}

	if pointer.Version == 0 {
		return errors.New("version must be greater than zero")
	}

	expectedKey := contact.NewKademliaIDFromData(
		[]byte(fmt.Sprintf(
			"%s:%s:latest",
			pointer.DomainName,
			pointer.PackageName,
		)),
	)

	if !pointer.Key().Equals(expectedKey) {
		return errors.New("latest pointer key mismatch")
	}

	if _, err := contact.ParseKademliaID(pointer.VersionRecordHash); err != nil {
		return fmt.Errorf("invalid version record hash: %w", err)
	}

	// 2. Verify the pointer's signature against the domain owner's key.
	pk, err := k.DNS.LookupPK(pointer.DomainName)
	if err != nil {
		return fmt.Errorf("DNS lookup failed: %w", err)
	}

	if !k.Signature.VerifyLatestRecord(pointer, pk) {
		return errors.New("invalid latest pointer signature")
	}

	// 3. Validate the referenced record and its complete history.
	head, err := k.validateVersionChain(
		pointer.VersionRecordHash,
		pointer.DomainName,
		pointer.PackageName,
	)
	if err != nil {
		return fmt.Errorf("invalid version history: %w", err)
	}

	if head.Version != pointer.Version {
		return errors.New("pointer version does not match version record")
	}

	// 4. With no local pointer, the valid chain back to genesis is enough
	// to validate the history we have received.
	if current == nil {
		return nil
	}

	// The existing pointer must belong to the same package.
	if current.DomainName != pointer.DomainName ||
		current.PackageName != pointer.PackageName {
		return errors.New("current pointer identity mismatch")
	}

	// 5. An exact repeat is an idempotent update, not a fork.
	if current.VersionRecordHash == pointer.VersionRecordHash &&
		current.Version == pointer.Version {
		return nil
	}

	// 6. Reject rollback and conflicting updates at the same version.
	if pointer.Version <= current.Version {
		return errors.New("rollback or conflicting version rejected")
	}

	// 7. The proposed chain must include our current head.
	// This allows catch-up when we have missed intermediate updates.
	if err := k.extendsCurrentHead(
		pointer.VersionRecordHash,
		current.VersionRecordHash,
		pointer.DomainName,
		pointer.PackageName,
	); err != nil {
		return fmt.Errorf("proposed update does not extend current history: %w", err)
	}

	return nil
}

func (k *Kademlia) extendsCurrentHead(
	proposedHash string,
	currentHash string,
	domain string,
	pkg string,
) error {
	hash := proposedHash

	for hash != zeroHash() {
		if hash == currentHash {
			return nil
		}

		record, err := k.GetVersionRecord(hash)
		if err != nil {
			return err
		}

		if record.DomainName != domain ||
			record.PackageName != pkg {
			return errors.New("version chain identity mismatch")
		}

		hash = record.PreviousVersionRecord
	}

	return errors.New("proposed version does not extend current history")
}

func zeroHash() string {
	return "0000000000000000000000000000000000000000000000000000000000000000"
}

func (k *Kademlia) validateVersionChain(
	headHash string,
	domain string,
	pkg string,
) (*VersionRecord, error) {
	if _, err := contact.ParseKademliaID(headHash); err != nil {
		return nil, fmt.Errorf("invalid head record hash: %w", err)
	}

	hash := headHash
	seen := make(map[string]bool)

	var head *VersionRecord
	var newer *VersionRecord
	var oldest *VersionRecord

	for hash != zeroHash() {
		// Reject cycles so malformed history cannot loop forever.
		if seen[hash] {
			return nil, errors.New("cycle detected in version history")
		}
		seen[hash] = true

		// LookupData verifies that the returned bytes hash to this key.
		// GetVersionRecord also verifies the record's signature.
		rec, err := k.GetVersionRecord(hash)
		if err != nil {
			return nil, fmt.Errorf(
				"could not retrieve version record %s: %w",
				hash, err,
			)
		}

		// Verify the record belongs to the expected package.
		if rec.DomainName != domain || rec.PackageName != pkg {
			return nil, errors.New("version record package identity mismatch")
		}

		// Ensure the decoded record's canonical hash matches the key.
		if rec.Hash().String() != hash {
			return nil, errors.New("version record hash mismatch")
		}

		if rec.Version == 0 {
			return nil, errors.New("version number must be greater than zero")
		}

		if head == nil {
			head = rec
		}

		// We're walking newest to oldest. Each predecessor must have
		// a strictly smaller version number than the record after it.
		if newer != nil && rec.Version >= newer.Version {
			return nil, errors.New(
				"version numbers must strictly increase along the chain",
			)
		}

		previous := rec.PreviousVersionRecord

		// Every non-genesis predecessor must be a valid Kademlia ID.
		if previous != zeroHash() {
			if _, err := contact.ParseKademliaID(previous); err != nil {
				return nil, fmt.Errorf(
					"invalid previous-record hash %q: %w",
					previous, err,
				)
			}
		}

		oldest = rec
		newer = rec
		hash = previous
	}

	if head == nil || oldest == nil {
		return nil, errors.New("version history is empty")
	}

	// The last record visited must be the genesis record.
	if oldest.PreviousVersionRecord != zeroHash() {
		return nil, errors.New("version history does not terminate at genesis")
	}

	return head, nil
}

func (k *Kademlia) PrintRT() {
	k.Routing.Mu.Lock()
	defer k.Routing.Mu.Unlock()

	buckets := k.Routing.GetBuckets()

	if len(buckets) == 0 {
		fmt.Println("Routing table is empty.")
		return
	}

	fmt.Println("Routing Table:")

	for i, bucket := range buckets {

		if bucket.Len() == 0 {
			continue
		}

		contacts := bucket.GetContacts()

		fmt.Printf("  bucket %d:\n", i)
		for _, c := range contacts {
			fmt.Printf("    %s: %s\n", c.Address, truncateId(*c.ID))
		}
	}
}

func truncateId(id contact.KademliaID) string {
	length := 8

	str := id.String()

	return str[0:length] + "..." + str[len(str)-length:]
}

package kademlia

import (
	"crypto/sha256"
	"d7024e/internal/kademlia/contact"
	"d7024e/internal/kademlia/datastore"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

const (
	versionRecordTag = "version-record"
	latestPointerTag = "latest-pointer"
)

// VersionRecord identifies one immutable version of a package.
type VersionRecord struct {
	Tag                   string `json:"tag"`
	DomainName            string `json:"domain-name"`
	PackageName           string `json:"package-name"`
	Version               uint64 `json:"version"`
	BlobHash              string `json:"blobHash"`
	PreviousVersionRecord string `json:"previous-version-record"`
	Signature             []byte `json:"sig"`
}

// LatestPointer identifies the head of a package's version-record chain.
type LatestPointer struct {
	Tag               string `json:"tag"`
	DomainName        string `json:"domain-name"`
	PackageName       string `json:"package-name"`
	Version           uint64 `json:"version"`
	VersionRecordHash string `json:"versionRecordHash"`
	Signature         []byte `json:"sig"`
}

func NewVersionRecord(domainName, packageName string, version uint64, blobHash, previousVersionRecord string, signature []byte) VersionRecord {
	return VersionRecord{
		Tag:                   versionRecordTag,
		DomainName:            domainName,
		PackageName:           packageName,
		Version:               version,
		BlobHash:              blobHash,
		PreviousVersionRecord: previousVersionRecord,
		Signature:             append([]byte(nil), signature...),
	}
}

func NewLatestPointer(domainName, packageName string, version uint64, versionRecordHash string, signature []byte) LatestPointer {
	return LatestPointer{
		Tag:               latestPointerTag,
		DomainName:        domainName,
		PackageName:       packageName,
		Version:           version,
		VersionRecordHash: versionRecordHash,
		Signature:         append([]byte(nil), signature...),
	}
}

type signingVersionRecord struct {
	Tag                   string `json:"tag"`
	DomainName            string `json:"domain-name"`
	PackageName           string `json:"package-name"`
	Version               uint64 `json:"version"`
	BlobHash              string `json:"blobHash"`
	PreviousVersionRecord string `json:"previous-version-record"`
}

type signingLatestPointer struct {
	Tag               string `json:"tag"`
	DomainName        string `json:"domain-name"`
	PackageName       string `json:"package-name"`
	Version           uint64 `json:"version"`
	VersionRecordHash string `json:"versionRecordHash"`
}

func (record VersionRecord) signingValue() signingVersionRecord {
	return signingVersionRecord{
		Tag:                   record.Tag,
		DomainName:            record.DomainName,
		PackageName:           record.PackageName,
		Version:               record.Version,
		BlobHash:              record.BlobHash,
		PreviousVersionRecord: record.PreviousVersionRecord,
	}
}

func (pointer LatestPointer) signingValue() signingLatestPointer {
	return signingLatestPointer{
		Tag:               pointer.Tag,
		DomainName:        pointer.DomainName,
		PackageName:       pointer.PackageName,
		Version:           pointer.Version,
		VersionRecordHash: pointer.VersionRecordHash,
	}
}

func (record VersionRecord) SigningBytes() []byte {
	data, _ := json.Marshal(record.signingValue())
	return data
}

func (pointer LatestPointer) SigningBytes() []byte {
	data, _ := json.Marshal(pointer.signingValue())
	return data
}

func (record VersionRecord) Serialize() []byte {
	data, _ := json.Marshal(record)
	return data
}

func (pointer LatestPointer) Serialize() []byte {
	data, _ := json.Marshal(pointer)
	return data
}

func (record VersionRecord) Hash() *contact.KademliaID {
	return contact.NewKademliaIDFromData(record.Serialize())
}

func (pointer LatestPointer) Hash() *contact.KademliaID {
	return contact.NewKademliaIDFromData(pointer.Serialize())
}

func (record VersionRecord) SigningHash() [32]byte {
	return sha256.Sum256(record.SigningBytes())
}

func (pointer LatestPointer) SigningHash() [32]byte {
	return sha256.Sum256(pointer.SigningBytes())
}

// use kademlia Store implementation instead since the data is immutable?
// TODO: Should be replicated over the DHT
func (record VersionRecord) Store(ds *datastore.DataStore) error {
	return ds.Put(record.Hash(), record.Serialize())
}

func (pointer LatestPointer) Key() *contact.KademliaID {
	return contact.NewKademliaIDFromData([]byte(pointer.DomainName + ":" + pointer.PackageName + ":latest"))
}

// TODO: Should be replicated over the DHT
func (pointer LatestPointer) Store(ds *datastore.DataStore) error {
	return ds.Put(pointer.Key(), pointer.Serialize())
}

func LoadVersionRecord(ds *datastore.DataStore, key *contact.KademliaID) (VersionRecord, error) {
	data, ok := ds.Get(key)
	if !ok {
		return VersionRecord{}, fmt.Errorf("kademlia: version record %s not found", key.String())
	}
	var record VersionRecord
	if err := json.Unmarshal(data, &record); err != nil {
		return VersionRecord{}, err
	}
	return record, nil
}

func LoadLatestPointer(ds *datastore.DataStore, key *contact.KademliaID) (LatestPointer, error) {
	data, ok := ds.Get(key)
	if !ok {
		return LatestPointer{}, fmt.Errorf("kademlia: latest pointer %s not found", key.String())
	}
	var pointer LatestPointer
	if err := json.Unmarshal(data, &pointer); err != nil {
		return LatestPointer{}, err
	}
	return pointer, nil
}

func (record VersionRecord) HashString() (string, error) {
	hash := record.Hash()
	return hex.EncodeToString(hash[:]), nil
}

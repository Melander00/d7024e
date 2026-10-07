package kademlia

import (
	c "crypto/ed25519"
)

type Signature interface {
	SignVersionRecord(rec VersionRecord) VersionRecord
	SignLatestRecord(rec LatestPointer) LatestPointer

	VerifyVersionRecord(rec VersionRecord, pk c.PublicKey) bool
	VerifyLatestRecord(rec LatestPointer, pk c.PublicKey) bool

	GetPublicKey() c.PublicKey
}

type ED25519 struct {
	sk c.PrivateKey
	Pk c.PublicKey
}

// generate / load key pairs
func NewEd25519Sign() (*ED25519, error) {
	pk, sk, err := c.GenerateKey(nil)

	if err != nil {
		return nil, err
	}

	return &ED25519{
		sk: sk,
		Pk: pk,
	}, nil
}

func (ed *ED25519) GetPublicKey() c.PublicKey {
	return ed.Pk
}

// Sign records
func (ed *ED25519) SignVersionRecord(versionRec VersionRecord) VersionRecord {
	hash := versionRec.SigningHash()
	sig := c.Sign(ed.sk, hash[:])
	versionRec.Signature = sig
	return versionRec
}

func (ed *ED25519) SignLatestRecord(versionRec LatestPointer) LatestPointer {
	hash := versionRec.SigningHash()
	sig := c.Sign(ed.sk, hash[:])
	versionRec.Signature = sig
	return versionRec
}

// Verify signatures
func (ed *ED25519) VerifyVersionRecord(rec VersionRecord, pk c.PublicKey) bool {
	hash := rec.SigningHash()
	sig := rec.Signature
	return c.Verify(pk, hash[:], sig)
}

func (ed *ED25519) VerifyLatestRecord(rec LatestPointer, pk c.PublicKey) bool {
	hash := rec.SigningHash()
	sig := rec.Signature
	return c.Verify(pk, hash[:], sig)
}

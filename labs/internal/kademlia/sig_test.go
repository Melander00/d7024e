package kademlia

import (
	"bytes"
	"crypto/ed25519"
	"testing"
)

func TestED25519SignAndVerifyVersionRecord(t *testing.T) {
	signer, err := NewEd25519Sign()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	record := NewVersionRecord(
		"example.com",
		"my-package",
		1,
		"blob-hash-123",
		"previous-record-hash",
		nil,
	)

	signed := signer.SignVersionRecord(record)

	if len(signed.Signature) == 0 {
		t.Fatal("expected version record to have a signature")
	}

	if !signer.VerifyVersionRecord(signed, signer.Pk) {
		t.Fatal("expected valid signature to verify")
	}
}

func TestED25519SignAndVerifyLatestPointer(t *testing.T) {
	signer, err := NewEd25519Sign()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	pointer := NewLatestPointer(
		"example.com",
		"my-package",
		1,
		"version-record-hash",
		nil,
	)

	signed := signer.SignLatestRecord(pointer)

	if len(signed.Signature) == 0 {
		t.Fatal("expected latest pointer to have a signature")
	}

	if !signer.VerifyLatestRecord(signed, signer.Pk) {
		t.Fatal("expected valid signature to verify")
	}
}

func TestED25519RejectsTamperedVersionRecord(t *testing.T) {
	signer, err := NewEd25519Sign()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	original := NewVersionRecord(
		"example.com",
		"my-package",
		1,
		"blob-hash-123",
		"previous-record-hash",
		nil,
	)

	signed := signer.SignVersionRecord(original)

	if !signer.VerifyVersionRecord(signed, signer.Pk) {
		t.Fatal("expected original record signature to verify")
	}

	tests := []struct {
		name   string
		tamper func(*VersionRecord)
	}{
		{
			name: "tag",
			tamper: func(r *VersionRecord) {
				r.Tag = "tampered-tag"
			},
		},
		{
			name: "domain name",
			tamper: func(r *VersionRecord) {
				r.DomainName = "attacker.example.com"
			},
		},
		{
			name: "package name",
			tamper: func(r *VersionRecord) {
				r.PackageName = "malicious-package"
			},
		},
		{
			name: "version",
			tamper: func(r *VersionRecord) {
				r.Version = 2
			},
		},
		{
			name: "blob hash",
			tamper: func(r *VersionRecord) {
				r.BlobHash = "tampered-blob-hash"
			},
		},
		{
			name: "previous version record",
			tamper: func(r *VersionRecord) {
				r.PreviousVersionRecord = "tampered-previous-record"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tampered := signed
			tt.tamper(&tampered)

			if signer.VerifyVersionRecord(tampered, signer.Pk) {
				t.Fatalf("expected tampered %s to fail verification", tt.name)
			}
		})
	}
}

func TestED25519RejectsTamperedLatestPointer(t *testing.T) {
	signer, err := NewEd25519Sign()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	original := NewLatestPointer(
		"example.com",
		"my-package",
		1,
		"version-record-hash",
		nil,
	)

	signed := signer.SignLatestRecord(original)

	if !signer.VerifyLatestRecord(signed, signer.Pk) {
		t.Fatal("expected original pointer signature to verify")
	}

	tests := []struct {
		name   string
		tamper func(*LatestPointer)
	}{
		{
			name: "tag",
			tamper: func(p *LatestPointer) {
				p.Tag = "tampered-tag"
			},
		},
		{
			name: "domain name",
			tamper: func(p *LatestPointer) {
				p.DomainName = "attacker.example.com"
			},
		},
		{
			name: "package name",
			tamper: func(p *LatestPointer) {
				p.PackageName = "malicious-package"
			},
		},
		{
			name: "version",
			tamper: func(p *LatestPointer) {
				p.Version = 2
			},
		},
		{
			name: "version record hash",
			tamper: func(p *LatestPointer) {
				p.VersionRecordHash = "tampered-record-hash"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tampered := signed
			tt.tamper(&tampered)

			if signer.VerifyLatestRecord(tampered, signer.Pk) {
				t.Fatalf("expected tampered %s to fail verification", tt.name)
			}
		})
	}
}

func TestED25519RejectsTamperedSignature(t *testing.T) {
	signer, err := NewEd25519Sign()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	record := NewVersionRecord(
		"example.com",
		"my-package",
		1,
		"blob-hash-123",
		"previous-record-hash",
		nil,
	)

	signed := signer.SignVersionRecord(record)

	if !signer.VerifyVersionRecord(signed, signer.Pk) {
		t.Fatal("expected original signature to verify")
	}

	signed.Signature[0] ^= 0xff

	if signer.VerifyVersionRecord(signed, signer.Pk) {
		t.Fatal("expected modified signature to fail verification")
	}
}

func TestED25519RejectsWrongPublicKey(t *testing.T) {
	signer, err := NewEd25519Sign()
	if err != nil {
		t.Fatalf("failed to generate first key pair: %v", err)
	}

	otherSigner, err := NewEd25519Sign()
	if err != nil {
		t.Fatalf("failed to generate second key pair: %v", err)
	}

	record := NewVersionRecord(
		"example.com",
		"my-package",
		1,
		"blob-hash-123",
		"previous-record-hash",
		nil,
	)

	signed := signer.SignVersionRecord(record)

	if signer.VerifyVersionRecord(signed, signer.Pk) == false {
		t.Fatal("expected signature to verify with signing public key")
	}

	if signer.VerifyVersionRecord(signed, otherSigner.Pk) {
		t.Fatal("expected signature to fail with a different public key")
	}
}

func TestED25519GeneratedKeyPairIsValid(t *testing.T) {
	signer, err := NewEd25519Sign()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	if len(signer.Pk) != ed25519.PublicKeySize {
		t.Fatalf(
			"unexpected public key size: got %d, want %d",
			len(signer.Pk),
			ed25519.PublicKeySize,
		)
	}

	if len(signer.sk) != ed25519.PrivateKeySize {
		t.Fatalf(
			"unexpected private key size: got %d, want %d",
			len(signer.sk),
			ed25519.PrivateKeySize,
		)
	}

	// The public key embedded in the private key should match Pk.
	privateKeyPublicKey := signer.sk.Public().(ed25519.PublicKey)

	if !bytes.Equal(privateKeyPublicKey, signer.Pk) {
		t.Fatal("public key does not match the public key derived from private key")
	}
}

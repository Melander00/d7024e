package dns

import (
	c "crypto/ed25519"
)

type DNS interface {
	LookupPK(domain string) (c.PublicKey, error)
	ClaimDomain(domain string, pk c.PublicKey) error
}

package dns

import (
	c "crypto/ed25519"
	"fmt"
	"sync"
)

type SimulatedDNS struct {
	mu      sync.RWMutex
	records map[string]c.PublicKey
}

func NewSimulatedDNS() *SimulatedDNS {
	return &SimulatedDNS{
		records: make(map[string]c.PublicKey),
	}
}

func (dns *SimulatedDNS) LookupPK(domain string) (c.PublicKey, error) {
	dns.mu.RLock()
	pk, ok := dns.records[domain]
	dns.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("domain not found")
	}
	return pk, nil
}

func (dns *SimulatedDNS) ClaimDomain(domain string, pk c.PublicKey) error {
	defer dns.mu.Unlock()
	dns.mu.Lock()

	_, ok := dns.records[domain]
	if ok {
		return fmt.Errorf("domain already claimed")
	}

	dns.records[domain] = pk

	return nil
}

package dns

import (
	"bytes"
	c "crypto/ed25519"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type RealDNS struct {
	dnsHost string
}

type LookupResponseBody struct {
	PK []byte `json:"pk"`
}

func NewRealDNS(host string) *RealDNS {
	return &RealDNS{
		dnsHost: host,
	}
}

func (dns *RealDNS) LookupPK(domain string) (c.PublicKey, error) {
	res, err := http.Get("http://" + dns.dnsHost + "/pk/" + domain)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DNS lookup failed: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	b := LookupResponseBody{}

	err = json.Unmarshal(body, &b)

	if err != nil {
		return nil, err
	}

	return c.PublicKey(b.PK), nil
}

func (dns *RealDNS) ClaimDomain(domain string, pk c.PublicKey) error {
	data := struct {
		Domain string      `json:"domain"`
		PK     c.PublicKey `json:"pk"`
	}{
		Domain: domain,
		PK:     pk,
	}

	body, err := json.Marshal(data)

	if err != nil {
		return err
	}

	res, err := http.Post(
		"http://"+dns.dnsHost+"/claim",
		"application/json",
		bytes.NewBuffer(body),
	)

	if err != nil {
		return err
	}

	if res.StatusCode == http.StatusOK {
		return nil
	}

	return errors.New("domain already claimed")
}

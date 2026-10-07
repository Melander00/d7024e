package dns

import (
	c "crypto/ed25519"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func testPublicKey() c.PublicKey {
	pk, _, err := c.GenerateKey(nil)
	if err != nil {
		panic(err)
	}
	return pk
}

func TestRealDNS_LookupPK(t *testing.T) {
	pk := testPublicKey()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		if r.URL.Path != "/pk/example.com" {
			t.Errorf("expected path /pk/example.com, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(LookupResponseBody{PK: pk})
	}))
	defer server.Close()

	dns := NewRealDNS(server.URL)

	got, err := dns.LookupPK("example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !c.PublicKey(got).Equal(pk) {
		t.Fatal("returned public key does not match expected key")
	}
}

func TestRealDNS_LookupPK_NonOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	dns := NewRealDNS(server.URL)

	got, err := dns.LookupPK("missing.example")
	if err == nil {
		t.Fatal("expected error")
	}

	if got != nil {
		t.Fatalf("expected nil public key, got %v", got)
	}
}

func TestRealDNS_LookupPK_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	dns := NewRealDNS(server.URL)

	got, err := dns.LookupPK("example.com")
	if err == nil {
		t.Fatal("expected JSON parsing error")
	}

	if got != nil {
		t.Fatalf("expected nil public key, got %v", got)
	}
}

func TestRealDNS_ClaimDomain(t *testing.T) {
	pk := testPublicKey()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if r.URL.Path != "/claim" {
			t.Errorf("expected path /claim, got %s", r.URL.Path)
		}

		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", got)
		}

		var body struct {
			Domain string      `json:"domain"`
			PK     c.PublicKey `json:"pk"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		if body.Domain != "example.com" {
			t.Errorf("expected domain example.com, got %q", body.Domain)
		}

		if !body.PK.Equal(pk) {
			t.Error("request contained incorrect public key")
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	dns := NewRealDNS(server.URL)

	if err := dns.ClaimDomain("example.com", pk); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRealDNS_ClaimDomain_AlreadyClaimed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "already claimed", http.StatusConflict)
	}))
	defer server.Close()

	dns := NewRealDNS(server.URL)

	err := dns.ClaimDomain("example.com", testPublicKey())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewSimulatedDNS(t *testing.T) {
	dns := NewSimulatedDNS()

	if dns == nil {
		t.Fatal("expected non-nil SimulatedDNS")
	}

	if dns.records == nil {
		t.Fatal("expected records map to be initialized")
	}

	if len(dns.records) != 0 {
		t.Fatalf("expected empty records map, got %d records", len(dns.records))
	}
}

func TestSimulatedDNS_LookupPK_NotFound(t *testing.T) {
	dns := NewSimulatedDNS()

	pk, err := dns.LookupPK("missing.example")
	if err == nil {
		t.Fatal("expected error")
	}

	if pk != nil {
		t.Fatalf("expected nil public key, got %v", pk)
	}
}

func TestSimulatedDNS_ClaimAndLookup(t *testing.T) {
	dns := NewSimulatedDNS()
	pk := testPublicKey()

	if err := dns.ClaimDomain("example.com", pk); err != nil {
		t.Fatalf("unexpected claim error: %v", err)
	}

	got, err := dns.LookupPK("example.com")
	if err != nil {
		t.Fatalf("unexpected lookup error: %v", err)
	}

	if !got.Equal(pk) {
		t.Fatal("lookup returned incorrect public key")
	}
}

func TestSimulatedDNS_ClaimDomain_Duplicate(t *testing.T) {
	dns := NewSimulatedDNS()

	firstPK := testPublicKey()
	secondPK := testPublicKey()

	if err := dns.ClaimDomain("example.com", firstPK); err != nil {
		t.Fatalf("unexpected first claim error: %v", err)
	}

	err := dns.ClaimDomain("example.com", secondPK)
	if err == nil {
		t.Fatal("expected duplicate claim error")
	}

	got, err := dns.LookupPK("example.com")
	if err != nil {
		t.Fatalf("unexpected lookup error: %v", err)
	}

	if !got.Equal(firstPK) {
		t.Fatal("duplicate claim replaced the original public key")
	}
}

func TestSimulatedDNS_MultipleDomains(t *testing.T) {
	dns := NewSimulatedDNS()

	domains := map[string]c.PublicKey{
		"one.example":   testPublicKey(),
		"two.example":   testPublicKey(),
		"three.example": testPublicKey(),
	}

	for domain, pk := range domains {
		if err := dns.ClaimDomain(domain, pk); err != nil {
			t.Fatalf("failed to claim %s: %v", domain, err)
		}
	}

	for domain, expected := range domains {
		got, err := dns.LookupPK(domain)
		if err != nil {
			t.Fatalf("failed to lookup %s: %v", domain, err)
		}

		if !got.Equal(expected) {
			t.Errorf("wrong public key for %s", domain)
		}
	}
}

func TestSimulatedDNS_ConcurrentClaims(t *testing.T) {
	dns := NewSimulatedDNS()
	pk := testPublicKey()

	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	errors := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			errors <- dns.ClaimDomain("example.com", pk)
		}()
	}

	wg.Wait()
	close(errors)

	successes := 0
	for err := range errors {
		if err == nil {
			successes++
		}
	}

	if successes != 1 {
		t.Fatalf("expected exactly one successful claim, got %d", successes)
	}

	got, err := dns.LookupPK("example.com")
	if err != nil {
		t.Fatalf("unexpected lookup error: %v", err)
	}

	if !got.Equal(pk) {
		t.Fatal("stored public key does not match claimed key")
	}
}

func TestSimulatedDNS_ConcurrentLookups(t *testing.T) {
	dns := NewSimulatedDNS()
	pk := testPublicKey()

	if err := dns.ClaimDomain("example.com", pk); err != nil {
		t.Fatalf("unexpected claim error: %v", err)
	}

	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			got, err := dns.LookupPK("example.com")
			if err != nil {
				t.Errorf("unexpected lookup error: %v", err)
				return
			}

			if !got.Equal(pk) {
				t.Error("lookup returned incorrect public key")
			}
		}()
	}

	wg.Wait()
}

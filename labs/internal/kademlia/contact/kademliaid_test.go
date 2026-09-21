package contact

import (
	"strings"
	"testing"
)

func TestParseKademliaID(t *T) {
	hex := "FFFFFFFF00000000000000000000000000000000000000000000000000000000"

	id, _ := ParseKademliaID(hex)

	if strings.ToLower(id.String()) != strings.ToLower(hex) {
		t.Fatalf("invalid ID %s should've been %s", id.String(), hex)
	}
}

func TestParseInvalidKademliaID(t *T) {
	// Too short
	_, err := ParseKademliaID("FFFFFFFF000000000000000000000000000000000000000000000000000000")
	if err == nil {
		t.Fatalf("parser didnt catch too short")
	}
	// Too long
	_, err = ParseKademliaID("FFFFFFFF0000000000000000000000000000000000000000000000000000000000")
	if err == nil {
		t.Fatalf("parser didnt catch too long")
	}
	// Invalid hex
	_, err = ParseKademliaID("ZZFFFFFF00000000000000000000000000000000000000000000000000000000")
	if err == nil {
		t.Fatalf("parser didnt catch invalid hex")
	}
}

func TestNewIDFromAddress(t *T) {
	addr := "127.0.0.1:10001"

	id, _ := NewKademliaIDFromAddress(addr)

	// Sha256 generated from outside golang
	realId, _ := ParseKademliaID("02b5b94b75a96720af96f713795941897f69cad9acafa3b2f8380089c7ee3c07")

	if !id.Equals(realId) {
		t.Fatalf("hash is wrong %s but should've been %s", id.String(), realId.String())
	}
}

func TestNewIDFromInvalidAddress(t *T) {
	addr := "some-faulty-address"

	_, err := NewKademliaIDFromAddress(addr)

	if err == nil {
		t.Fatalf("didnt catch invalid address")
	}
}

func TestNewIDFromData(t *T) {
	expected := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"

	first := NewKademliaIDFromData([]byte("hello"))
	second := NewKademliaIDFromData([]byte("hello"))

	if first.String() != expected {
		t.Fatalf("hash is wrong %s but should've been %s", first.String(), expected)
	}

	if !first.Equals(second) {
		t.Fatalf("identical data should produce identical IDs: %s | %s", first.String(), second.String())
	}

	if len(first.String()) != IDLength*2 {
		t.Fatalf("expected ID string length %d, got %d", IDLength*2, len(first.String()))
	}
}

func TestParseMalformedKademliaID(t *T) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "too short", key: "00"},
		{name: "too long", key: "000000000000000000000000000000000000000000000000000000000000000000"},
		{name: "invalid hex", key: "zz00000000000000000000000000000000000000000000000000000000000000"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *T) {
			if _, err := ParseKademliaID(test.key); err == nil {
				t.Fatalf("expected malformed key %q to be rejected", test.key)
			}
		})
	}
}

func TestIDLess(t *T) {
	id1, _ := ParseKademliaID("FFFFFFFF00000000000000000000000000000000000000000000000000000000")
	id2, _ := ParseKademliaID("1111111100000000000000000000000000000000000000000000000000000000")

	if !id2.Less(id1) {
		t.Fatalf("%s was less than %s", id1.String(), id2.String())
	}
}

func TestIDEquals(t *T) {

	hex := "FFFFFFFF00000000000000000000000000000000000000000000000000000000"

	id1, _ := ParseKademliaID(hex)
	id2, _ := ParseKademliaID(hex)

	if !id1.Equals(id2) {
		t.Fatalf("invalid identical ids %s | %s", id1.String(), id2.String())
	}
}

func TestIDDistance(t *T) {
	id1, _ := ParseKademliaID("FFFFFFFF00000000000000000000000000000000000000000000000000000000")
	id2, _ := ParseKademliaID("1111111100000000000000000000000000000000000000000000000000000000")

	d, _ := ParseKademliaID("EEEEEEEE00000000000000000000000000000000000000000000000000000000")

	if !id1.CalcDistance(id2).Equals(d) {
		t.Fatalf("distance calculation is wrong, got %s but should be %s", id1.CalcDistance(id2), d)
	}

	ab := id1.CalcDistance(id2)
	ba := id2.CalcDistance(id1)

	if !ab.Equals(ba) {
		t.Fatalf("a XOR b should equal b XOR a, %s | %s", ab, ba)
	}

	zeroD, _ := ParseKademliaID("0000000000000000000000000000000000000000000000000000000000000000")

	if !id1.CalcDistance(id1).Equals(zeroD) {
		t.Fatalf("a XOR a should be zero")
	}

}

func TestNewRandomKademliaID(t *testing.T) {
	id := NewRandomKademliaID()
	if id == nil {
		t.Fatal("expected a generated Kademlia ID")
	}

	if id.String() == "0000000000000000000000000000000000000000000000000000000000000000" {
		t.Fatal("expected generated Kademlia ID to contain random data")
	}

	if len(id.String()) != IDLength*2 {
		t.Fatalf("expected ID string length %d, got %d", IDLength*2, len(id.String()))
	}
}

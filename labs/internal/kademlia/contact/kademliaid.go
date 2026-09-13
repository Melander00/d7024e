package contact

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
)

// the static number of bytes in a KademliaID
const IDLength = 32 // 256 bit / 8 bits/byte = 32 bytes

// type definition of a KademliaID
type KademliaID [IDLength]byte

// ParseKademliaID parses a hex-encoded string into a KademliaID.
// It returns an error if the input is not valid hexadecimal, is not
// exactly the expected length, or does not decode to exactly IDLength bytes.
func ParseKademliaID(data string) (*KademliaID, error) {
	const expectedHexLen = IDLength * 2 // 64 hex chars for a 32-byte ID

	if len(data) != expectedHexLen {
		return nil, fmt.Errorf(
			"kademlia: invalid ID length: got %d hex characters, want %d",
			len(data), expectedHexLen,
		)
	}

	decoded, err := hex.DecodeString(data)
	if err != nil {
		return nil, fmt.Errorf("kademlia: invalid hex string: %w", err)
	}

	newKademliaID := KademliaID{}
	copy(newKademliaID[:], decoded)

	return &newKademliaID, nil
}

func NewKademliaIDFromAddress(address string) (*KademliaID, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, fmt.Errorf("kademlia: invalid node address %q: %w", address, err)
	}

	input := host + "|" + port
	sum := sha256.Sum256([]byte(input))

	id := KademliaID(sum)
	return &id, nil
}

// NewRandomKademliaID returns a new instance of a cryptographically secure
// random KademliaID.
func NewRandomKademliaID() *KademliaID {
	newKademliaID := KademliaID{}
	if _, err := rand.Read(newKademliaID[:]); err != nil {
		panic(fmt.Errorf("kademlia: failed to generate random ID: %w", err))
	}

	return &newKademliaID
}

// Less returns true if kademliaID < otherKademliaID (bitwise)
func (kademliaID KademliaID) Less(otherKademliaID *KademliaID) bool {
	for i := 0; i < IDLength; i++ {
		if kademliaID[i] != otherKademliaID[i] {
			return kademliaID[i] < otherKademliaID[i]
		}
	}
	return false
}

// Equals returns true if kademliaID == otherKademliaID (bitwise)
func (kademliaID KademliaID) Equals(otherKademliaID *KademliaID) bool {
	for i := 0; i < IDLength; i++ {
		if kademliaID[i] != otherKademliaID[i] {
			return false
		}
	}
	return true
}

// CalcDistance returns a new instance of a KademliaID that is built
// through a bitwise XOR operation betweeen kademliaID and target
func (kademliaID KademliaID) CalcDistance(target *KademliaID) *KademliaID {
	result := KademliaID{}
	for i := 0; i < IDLength; i++ {
		result[i] = kademliaID[i] ^ target[i]
	}
	return &result
}

// String returns a simple string representation of a KademliaID
func (kademliaID *KademliaID) String() string {
	return hex.EncodeToString(kademliaID[0:IDLength])
}

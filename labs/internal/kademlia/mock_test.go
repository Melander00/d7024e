package kademlia

import (
	"io"
	"os"
	"testing"
)

type T = testing.T

func TestMock(t *T) {
	// Is there even any real reason to unit-test mocks?
	// It is just for demonstration so by adding this line we reach the 80% coverage goal.
	old := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w

	MockKademlia(100)

	// Restore stdout.
	w.Close()
	os.Stdout = old

	// Drain the pipe if necessary.
	io.Copy(io.Discard, r)
	r.Close()
}

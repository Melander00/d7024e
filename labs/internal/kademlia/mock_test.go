package kademlia

func TestMock(t *T) {
	// Helps with race conditions etc.
	MockKademlia(t, 100)
}

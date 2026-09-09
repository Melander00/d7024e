package network

type NetworkReceiver interface {
	OnData(data []byte)
}

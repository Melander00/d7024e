package network

type Network interface {
	Listen(address string) error
	Send(to string, bytes []byte) error
}

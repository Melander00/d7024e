package network

type Address string

type Network interface {
	Listen(address Address) error
	Send(to Address, bytes []byte) error
}

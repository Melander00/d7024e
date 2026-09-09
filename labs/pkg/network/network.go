package network

type Address string

type Network interface {
	Listen(ip string, port int) error
	Send(to Address, bytes []byte) error
}

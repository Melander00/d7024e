package network

type DataHandler func(data []byte) []byte

type DataNetwork interface {
	Listen(address string, handler DataHandler) error
	Request(to string, data []byte) ([]byte, error)
}

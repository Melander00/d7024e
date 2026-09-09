package network

type SimulatedNetwork struct {
	receiver NetworkReceiver
}

func (net *SimulatedNetwork) Listen(ip string, port int) error {
	// TODO

	return nil
}

func (net *SimulatedNetwork) Send(to string, bytes []byte) error {
	// TODO

	return nil
}

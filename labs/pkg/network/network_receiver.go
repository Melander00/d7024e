package network

type NetworkReceiver interface {
	OnData(data []byte)
	Read(out chan []byte)
}

type ChannelNetworkReceiver struct {
	channel chan []byte
}

func NewChannelNetworkReceiver(size int) *ChannelNetworkReceiver {
	return &ChannelNetworkReceiver{
		channel: make(chan []byte, size),
	}
}

func (rec *ChannelNetworkReceiver) OnData(data []byte) {
	rec.channel <- data
}

func (rec *ChannelNetworkReceiver) Read(out chan []byte) {
	for data := range rec.channel {
		out <- data
	}
}

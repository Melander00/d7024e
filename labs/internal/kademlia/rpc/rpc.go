package rpc

import (
	"d7024e/pkg/network"
	"encoding/json"
	"fmt"
	"sync"
	"time"
	"uuid"
)

type MessageType string

const (
	PING       MessageType = "PING"
	PONG       MessageType = "PONG"
	FIND_NODE  MessageType = "FIND_NODE"
	FIND_VALUE MessageType = "FIND_VALUE"
	STORE      MessageType = "STORE"
)

type Message struct {
	Type      MessageType
	RequestID string
	Sender    network.Address
	Value     any
}

func marshalMessage(msg Message) []byte {
	b, _ := json.Marshal(msg)

	return b
}

func unmarshalMessage(data []byte) Message {
	var m Message

	json.Unmarshal(data, &m)

	return m
}

type RPC struct {
	network     network.Network
	me          network.Address
	pending     map[string]Request
	readChannel chan []byte
	mu          sync.RWMutex
}

type Request struct {
	// createdAt time.Time
	channel chan Message
	timeout time.Duration
}

func CreateRpc(receiver network.NetworkReceiver, net network.Network, me network.Address) *RPC {

	rpc := &RPC{
		network:     net,
		me:          me,
		readChannel: make(chan []byte, 5),
		pending:     make(map[string]Request),
	}

	go receiver.Read(rpc.readChannel)
	go rpc.Read()

	return rpc
}

func (rpc *RPC) Read() {
	for data := range rpc.readChannel {
		// data := <-rpc.readChannel
		msg := unmarshalMessage(data)

		rpc.mu.RLock()
		req, exists := rpc.pending[msg.RequestID]
		rpc.mu.RUnlock()

		if exists {
			req.channel <- msg
			rpc.mu.Lock()
			delete(rpc.pending, msg.RequestID) // Clean-up
			rpc.mu.Unlock()
		} else {

			// If there is no listener it is a request and not response
			// So we need to handle the request
		}
	}
}

func (rpc *RPC) Ping(to network.Address) (Message, error) {
	msg := rpc.createMessage(PING)
	req := rpc.createRequest(msg.RequestID)

	data := marshalMessage(msg)
	rpc.network.Send(to, data) // TODO: Error handling

	select {
	case res := <-req.channel:
		return res, nil
	case <-time.After(req.timeout):
		return Message{}, &ErrTimeout{}
	}
}

func (rpc *RPC) Store(to network.Address, key string, value []byte) {

}

func (rpc *RPC) FindNode(to network.Address, id string) {

}

func (rpc *RPC) FindValue(to network.Address, id string) {

}

func (rpc *RPC) createMessage(Type MessageType) Message {
	msg := *&Message{
		RequestID: uuid.NewV7().String(),
		Sender:    rpc.me,
		Type:      Type,
	}

	return msg
}

func (rpc *RPC) createRequest(requestID string) Request {
	req := Request{
		// createdAt: time.Now().Unix(),
		channel: make(chan Message, 1),
		timeout: 5 * time.Second,
	}

	rpc.mu.RLock()
	rpc.pending[requestID] = req
	rpc.mu.RUnlock()

	return req
}

type ErrTimeout struct {
}

func (e *ErrTimeout) Error() string {
	return fmt.Sprintf("Message timed out")
}

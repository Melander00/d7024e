package rpc

import (
	"d7024e/pkg/network"
	"encoding/json"
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
	network network.Network
	me      network.Address
}

func CreateRpc(net network.Network, me network.Address) *RPC {
	return &RPC{
		network: net,
		me:      me,
	}
}

func (rpc *RPC) OnData(data []byte) {
	// Decode data into JSON message

	// Find the channel associated with requestId
}

func (rpc *RPC) Ping(to network.Address) (Message, error) {
	msg := rpc.createMessage()
	msg.Type = PING
	data := marshalMessage(msg)
	rpc.network.Send(to, data)

	// Create a channel
	// Add channel to pending-list
	// Wait on that channel
	// Return the message

	// We keep this synchronous and enforce callers to make sure it uses valid async behaviour.

	return msg, nil // <-- placeholder
}

func (rpc *RPC) Store(to network.Address, key string, value []byte) {

}

func (rpc *RPC) FindNode(to network.Address, id string) {

}

func (rpc *RPC) FindValue(to network.Address, id string) {

}

func (rpc *RPC) createMessage() Message {
	msg := *&Message{
		RequestID: uuid.NewV7().String(),
		Sender:    rpc.me,
	}

	return msg
}

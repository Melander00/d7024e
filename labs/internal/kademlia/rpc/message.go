package rpc

import (
	"d7024e/internal/kademlia/contact"
	"d7024e/pkg/network"
	"encoding/json"
)

type MessageType string

const (
	PING                MessageType = "PING"
	PONG                MessageType = "PONG"
	FIND_NODE           MessageType = "FIND_NODE"
	FIND_NODE_RESPONSE  MessageType = "FIND_NODE_RESPONSE"
	FIND_VALUE          MessageType = "FIND_VALUE"
	FIND_VALUE_RESPONSE MessageType = "FIND_VALUE_RESPONSE"
	STORE               MessageType = "STORE"
	STORE_RESPONSE      MessageType = "STORE_RESPONSE"
)

type MessageValue struct {
	Hash    string
	Key     string
	Value   []byte
	Nodes   []contact.Contact
	Success bool
}

type Message struct {
	Type      MessageType
	RequestID string
	Sender    network.Address
	Value     MessageValue
}

func FindNodeRequest(hash string) MessageValue {
	return MessageValue{
		Hash: hash,
	}
}

func FindValueRequest(hash string) MessageValue {
	return MessageValue{
		Hash: hash,
	}
}

func StoreRequest(key string, value []byte) MessageValue {
	return MessageValue{
		Key:   key,
		Value: value,
	}
}

func FindValueResponse(data []byte) MessageValue {
	return MessageValue{
		Value: data,
	}
}

func FindNodeResponse(nodes []contact.Contact) MessageValue {
	return MessageValue{
		Nodes: nodes,
	}
}

func StoreResponse(success bool) MessageValue {
	return MessageValue{
		Success: success,
	}
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

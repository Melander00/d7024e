package rpc

import (
	"d7024e/internal/kademlia/contact"
	"d7024e/pkg/network"
	"sync"
	"time"
	"uuid"
)

type RPC struct {
	network          network.Network
	Me               network.Address
	pending          map[string]*Request
	readChannel      chan []byte
	mu               sync.RWMutex
	timeout          time.Duration
	retries          int
	pingHandler      func()
	findNodeHandler  func(hash string) []contact.Contact
	findValueHandler func(hash string) ([]contact.Contact, []byte, bool)
	storeHandler     func(key string, value []byte) bool
}

type Request struct {
	// createdAt time.Time
	channel chan Message
	timeout time.Duration
	retries int
	handled bool
}

func CreateRpc(receiver network.NetworkReceiver, net network.Network, me network.Address) *RPC {

	rpc := &RPC{
		network:     net,
		Me:          me,
		readChannel: make(chan []byte, 5),
		pending:     make(map[string]*Request),
		timeout:     5 * time.Second,
		retries:     5,
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

		if exists && req.handled == false {
			req.channel <- msg
			rpc.mu.Lock()
			// delete(rpc.pending, msg.RequestID) // Clean-up
			// With retries we dont want to delete the req handler but rather check if it already has been handled.
			// This means we need a garbage cleaner for this. TODO: implement a gc that checks how long ago it was created and if it was over 2 minutes we remove it.
			req.handled = true
			rpc.mu.Unlock()
		} else {
			go rpc.handleRequest(msg)
		}
	}
}

func (rpc *RPC) handleRequest(msg Message) {
	var res Message
	switch msg.Type {
	case PING:
		if rpc.pingHandler == nil {
			break
		}

		rpc.pingHandler()
		res = rpc.createMessage(PONG)

		break

	case FIND_NODE:
		if rpc.findNodeHandler == nil {
			break
		}

		data := rpc.findNodeHandler(msg.Value.Hash)
		res = rpc.createMessage(FIND_NODE_RESPONSE)
		res.Value = FindNodeResponse(data)

		break
	case FIND_VALUE:
		if rpc.findValueHandler == nil {
			break
		}

		candidates, data, hasValue := rpc.findValueHandler(msg.Value.Hash)
		res = rpc.createMessage(FIND_VALUE_RESPONSE)
		if hasValue {
			res.Value = FindValueResponse(data)
		} else {
			res.Value = FindNodeResponse(candidates)
		}

		break

	case STORE:
		if rpc.storeHandler == nil {
			break
		}

		data := rpc.storeHandler(msg.Value.Key, msg.Value.Value)
		res = rpc.createMessage(STORE_RESPONSE)
		res.Value = StoreResponse(data)

		break
	default:
		return
	}

	res.RequestID = msg.RequestID

	rpc.respond(msg.Sender, marshalMessage(res))
}

func (rpc *RPC) send(to network.Address, req *Request, data []byte, tryNr int) (*Message, error) {
	error := rpc.network.Send(to, data) // TODO: Error handling
	if error != nil {
		return nil, error
	}

	timer := time.NewTimer(req.timeout)
	defer timer.Stop()

	select {
	case res := <-req.channel:
		return &res, nil

	case <-timer.C:
		if tryNr >= req.retries {
			return &Message{}, &ErrTimeout{}
		}
		return rpc.send(to, req, data, tryNr+1)
	}
}

func (rpc *RPC) respond(to network.Address, data []byte) {
	rpc.network.Send(to, data)
}

func (rpc *RPC) Ping(to network.Address) (*Message, error) {
	msg := rpc.createMessage(PING)
	req := rpc.createRequest(msg.RequestID)

	data := marshalMessage(msg)
	return rpc.send(to, req, data, 1)
}

func (rpc *RPC) Store(to network.Address, key string, value []byte) (*Message, error) {
	msg := rpc.createMessage(STORE)
	req := rpc.createRequest(msg.RequestID)
	msg.Value = StoreRequest(key, value)

	data := marshalMessage(msg)
	return rpc.send(to, req, data, 1)

}

func (rpc *RPC) FindNode(to network.Address, id string) (*Message, error) {
	msg := rpc.createMessage(FIND_NODE)
	req := rpc.createRequest(msg.RequestID)
	msg.Value = FindNodeRequest(id)

	data := marshalMessage(msg)
	return rpc.send(to, req, data, 1)
}

func (rpc *RPC) FindValue(to network.Address, id string) (*Message, error) {
	msg := rpc.createMessage(FIND_VALUE)
	req := rpc.createRequest(msg.RequestID)
	msg.Value = FindValueRequest(id)

	data := marshalMessage(msg)
	return rpc.send(to, req, data, 1)
}

func (rpc *RPC) SetTimeout(timeout time.Duration) {
	rpc.timeout = timeout
}

func (rpc *RPC) SetRetries(retries int) {
	rpc.retries = retries
}

func (rpc *RPC) createMessage(Type MessageType) Message {
	msg := *&Message{
		RequestID: uuid.NewV7().String(),
		Sender:    rpc.Me,
		Type:      Type,
	}

	return msg
}

func (rpc *RPC) createRequest(requestID string) *Request {
	req := &Request{
		// createdAt: time.Now().Unix(),
		channel: make(chan Message, 1),
		timeout: rpc.timeout,
		retries: rpc.retries,
	}

	rpc.mu.Lock()
	rpc.pending[requestID] = req
	rpc.mu.Unlock()

	return req
}

/*

	pingHandler      func()
	findNodeHandler  func(hash string) []network.Address
	findValueHandler func(hash string) []byte
	storeHandler     func(key string, value []byte) bool

*/

func (rpc *RPC) SetPingHandler(handler func()) {
	rpc.pingHandler = handler
}

func (rpc *RPC) SetFindNodeHandler(handler func(hash string) []contact.Contact) {
	rpc.findNodeHandler = handler
}

func (rpc *RPC) SetFindValueHandler(handler func(hash string) ([]contact.Contact, []byte, bool)) {
	rpc.findValueHandler = handler
}

func (rpc *RPC) SetStoreHandler(handler func(key string, value []byte) bool) {
	rpc.storeHandler = handler
}

package rpc

import (
	"d7024e/internal/kademlia/contact"
	"d7024e/pkg/network"
	"errors"
	"sync"
	"time"
	"uuid"
)

type RPC struct {
	network          network.Network
	dataNetwork      network.DataNetwork
	Me               contact.Contact
	pending          map[string]*Request
	readChannel      chan []byte
	mu               sync.RWMutex
	timeout          time.Duration
	retries          int
	pingHandler      func(client contact.Contact)
	findNodeHandler  func(client contact.Contact, hash string) ([]contact.Contact, error)
	findValueHandler func(client contact.Contact, hash string) ([]contact.Contact, []byte, bool, error)
	storeHandler     func(client contact.Contact, key string, value []byte) bool
}

type Request struct {
	createdAt time.Time
	channel   chan Message
	timeout   time.Duration
	retries   int
	handled   bool
}

func CreateRpc(receiver network.NetworkReceiver, net network.Network, me contact.Contact) *RPC {

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
	go startGarbageCleaner(rpc, 5*time.Minute, 2*time.Minute)

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

		rpc.pingHandler(msg.Sender)
		res = rpc.createMessage(PONG)

		break

	case FIND_NODE:

		contacts, err :=
			rpc.findNodeHandler(
				msg.Sender,
				msg.Value.Hash,
			)

		res = rpc.createMessage(FIND_NODE_RESPONSE)

		if err != nil {
			res.Error = err.Error()
		} else {
			res.Value = FindNodeResponse(contacts)
		}
		break

	case FIND_VALUE:

		candidates, _, hasValue, err :=
			rpc.findValueHandler(
				msg.Sender,
				msg.Value.Hash,
			)

		res = rpc.createMessage(FIND_VALUE_RESPONSE)

		if err != nil {
			res.Error = err.Error()
			break
		}

		if hasValue {
			// Important:
			// Tell client we have it,
			// but don't put data in UDP response.
			res.Value.HasValue = true
		} else {
			res.Value = FindNodeResponse(candidates)
		}

		break

	default:
		return
	}

	res.RequestID = msg.RequestID

	rpc.respond(msg.Sender, marshalMessage(res))
}

func (rpc *RPC) send(to contact.Contact, req *Request, data []byte, tryNr int) (*Message, error) {
	error := rpc.network.Send(to.Address, data) // TODO: Error handling
	if error != nil {
		return nil, error
	}

	timer := time.NewTimer(req.timeout)
	defer timer.Stop()

	select {

	case res := <-req.channel:

		if res.Error != "" {
			return &res, errors.New(res.Error)
		}

		return &res, nil

	case <-timer.C:
		if tryNr >= req.retries {
			return &Message{}, &ErrTimeout{}
		}
		return rpc.send(to, req, data, tryNr+1)
	}
}

func (rpc *RPC) respond(to contact.Contact, data []byte) {
	rpc.network.Send(to.Address, data)
}

func (rpc *RPC) Ping(to contact.Contact) (*Message, error) {
	msg := rpc.createMessage(PING)
	req := rpc.createRequest(msg.RequestID)

	data := marshalMessage(msg)
	return rpc.send(to, req, data, 1)
}

func (rpc *RPC) Store(
	to contact.Contact,
	key string,
	value []byte,
) (*Message, error) {

	dataReq := DataRequest{
		Type:   DATA_STORE,
		Sender: rpc.Me,
		Key:    key,
		Value:  value,
	}

	raw, err := rpc.dataNetwork.Request(
		to.Address,
		marshalDataRequest(dataReq),
	)

	if err != nil {
		return nil, err
	}

	dataRes, err := unmarshalDataResponse(raw)
	if err != nil {
		return nil, err
	}

	if !dataRes.OK {
		return nil, errors.New(dataRes.Error)
	}

	// Keep existing RPC.Store return type for now.
	res := rpc.createMessage(STORE_RESPONSE)
	res.Value = StoreResponse(true)

	return &res, nil
}

func (rpc *RPC) FindNode(to contact.Contact, id string) (*Message, error) {
	msg := rpc.createMessage(FIND_NODE)
	req := rpc.createRequest(msg.RequestID)
	msg.Value = FindNodeRequest(id)

	data := marshalMessage(msg)
	return rpc.send(to, req, data, 1)
}

func (rpc *RPC) FindValue(
	to contact.Contact,
	id string,
) (*Message, error) {

	// Existing UDP request.
	msg := rpc.createMessage(FIND_VALUE)
	req := rpc.createRequest(msg.RequestID)
	msg.Value = FindValueRequest(id)

	res, err := rpc.send(
		to,
		req,
		marshalMessage(msg),
		1,
	)

	if err != nil {
		return nil, err
	}

	if !res.Value.HasValue {
		return res, nil
	}

	// Value exists remotely.
	// Fetch actual bytes through reliable data plane.
	dataReq := DataRequest{
		Type:   DATA_GET,
		Sender: rpc.Me,
		Key:    id,
	}

	raw, err := rpc.dataNetwork.Request(
		to.Address,
		marshalDataRequest(dataReq),
	)

	if err != nil {
		return nil, err
	}

	dataRes, err := unmarshalDataResponse(raw)
	if err != nil {
		return nil, err
	}

	if !dataRes.OK {
		return nil, errors.New(dataRes.Error)
	}

	// Verify that the received value actually matches the requested key.
	requestedID, err := contact.ParseKademliaID(id)
	if err != nil {
		return nil, err
	}

	receivedID := contact.NewKademliaIDFromData(dataRes.Value)
	if !requestedID.Equals(receivedID) {
		return nil, errors.New("received value does not match requested key")
	}

	// Preserve current API for Kademlia.
	res.Value.Value = dataRes.Value

	return res, nil
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
		createdAt: time.Now(),
		channel:   make(chan Message, 1),
		timeout:   rpc.timeout,
		retries:   rpc.retries,
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

func (rpc *RPC) SetPingHandler(handler func(client contact.Contact)) {
	rpc.pingHandler = handler
}

func (rpc *RPC) SetFindNodeHandler(
	handler func(
		client contact.Contact,
		hash string,
	) ([]contact.Contact, error),
) {
	rpc.findNodeHandler = handler
}

func (rpc *RPC) SetFindValueHandler(handler func(client contact.Contact, hash string) ([]contact.Contact, []byte, bool, error)) {
	rpc.findValueHandler = handler
}

func (rpc *RPC) SetStoreHandler(handler func(client contact.Contact, key string, value []byte) bool) {
	rpc.storeHandler = handler
}

// Longer-term, constructor injection would be cleaner???
func (rpc *RPC) SetDataNetwork(net network.DataNetwork) {
	rpc.dataNetwork = net
}

func (rpc *RPC) StartDataPlane() error {
	if rpc.dataNetwork == nil {
		return errors.New("data network is not configured")
	}

	return rpc.dataNetwork.Listen(
		rpc.Me.Address,
		rpc.handleDataRequest,
	)
}

// handleDataRequest processes incoming data requests (GET or STORE) and returns the appropriate response.
func (rpc *RPC) handleDataRequest(data []byte) []byte {

	req, err := unmarshalDataRequest(data)
	if err != nil {
		return marshalDataResponse(DataResponse{
			OK:    false,
			Error: err.Error(),
		})
	}

	switch req.Type {

	case DATA_GET:
		candidates, value, found, err :=
			rpc.findValueHandler(
				req.Sender,
				req.Key,
			)

		_ = candidates

		if err != nil {
			return marshalDataResponse(DataResponse{
				OK:    false,
				Error: err.Error(),
			})
		}

		if !found {
			return marshalDataResponse(DataResponse{
				OK:    false,
				Error: "value not found",
			})
		}

		return marshalDataResponse(DataResponse{
			OK:    true,
			Value: value,
		})

	case DATA_STORE:
		if rpc.storeHandler == nil {
			return marshalDataResponse(DataResponse{
				OK:    false,
				Error: "store handler is not configured",
			})
		}

		ok := rpc.storeHandler(
			req.Sender,
			req.Key,
			req.Value,
		)

		return marshalDataResponse(DataResponse{
			OK: ok,
		})
	}

	return marshalDataResponse(DataResponse{
		OK:    false,
		Error: "unknown data request",
	})
}

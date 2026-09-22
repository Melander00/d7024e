package rpc

import (
	"d7024e/internal/kademlia/contact"
	"encoding/json"
)

type DataRequestType string

const (
	DATA_GET   DataRequestType = "GET"
	DATA_STORE DataRequestType = "STORE"
)

type DataRequest struct {
	Type   DataRequestType
	Sender contact.Contact
	Key    string
	Value  []byte
}

type DataResponse struct {
	Value []byte
	OK    bool
	Error string
}

func marshalDataRequest(req DataRequest) []byte {
	data, _ := json.Marshal(req)
	return data
}

func unmarshalDataRequest(data []byte) (DataRequest, error) {
	var req DataRequest
	err := json.Unmarshal(data, &req)
	return req, err
}

func marshalDataResponse(res DataResponse) []byte {
	data, _ := json.Marshal(res)
	return data
}

func unmarshalDataResponse(data []byte) (DataResponse, error) {
	var res DataResponse
	err := json.Unmarshal(data, &res)
	return res, err
}

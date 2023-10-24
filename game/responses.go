package game

import (
	"google.golang.org/protobuf/reflect/protoreflect"
)

type ClientResponse struct {
	Payload protoreflect.ProtoMessage
	Ignore  map[uint64]bool
}

func CreateEvent(p map[uint64]bool, d protoreflect.ProtoMessage) ClientResponse {
	return ClientResponse{
		Payload: d,
		Ignore:  p,
	}
}

package core

import (
	"minik8s/pkg/api/types"
)

type ApiObjectType string

// These are the valid statuses of pods.
const (
	PodObjectType     ApiObjectType = "Pod"
	ServiceObjectType ApiObjectType = "Service"
	NodeObjectType    ApiObjectType = "Node"
)

type IApiObject interface {
	SetUID(uid types.UID)
	GetUID() types.UID
	JsonUnmarshal(data []byte) error
	JsonMarshal() ([]byte, error)
	JsonUnmarshalStatus(data []byte) error
	JsonMarshalStatus() ([]byte, error)
	SetStatus(s IApiObjectStatus) bool
	GetStatus() IApiObjectStatus
}

type IApiObjectStatus interface {
	JsonUnmarshal(data []byte) error
	JsonMarshal() ([]byte, error)
}

func CreateApiObject(ty ApiObjectType) IApiObject {
	switch ty {
	case PodObjectType:
		return &Pod{}
	case ServiceObjectType:
		return &Service{}
	case NodeObjectType:
		return &Node{}
	default:
	}
	return nil
}

func CreateApiObjectStatus(ty ApiObjectType) IApiObjectStatus {
	switch ty {
	case PodObjectType:
		return &PodStatus{}
	case ServiceObjectType:
		return &ServiceStatus{}
	case NodeObjectType:
		return &NodeStatus{}
	default:
	}
	return nil
}

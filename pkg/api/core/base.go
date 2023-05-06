package core

import (
	"fmt"
	"minik8s/pkg/api"
	"minik8s/pkg/api/types"
)

type ApiObjectType string

// These are the valid ApiObjectType.
const (
	ErrorObjectType   ApiObjectType = "Error"
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
	GetResourceVersion() string
	SetResourceVersion(version string)
	CreateFromEtcdString(str string) error
}

type IApiObjectList interface {
	JsonUnmarshal(data []byte) error
	JsonMarshal() ([]byte, error)
	AddItemFromStr(objectStr string) error
	AppendItemsFromStr(objectStrs []string) error
	GetItems() any
	GetIApiObjectArr() []IApiObject
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
	case ErrorObjectType:
		return &ErrorApiObject{}
	default:
		panic(fmt.Sprintf("No ApiObjectType %v", ty))
	}
	return nil
}

func CreateApiObjectList(ty ApiObjectType) IApiObjectList {
	switch ty {
	case PodObjectType:
		return &PodList{}
	case ServiceObjectType:
		return &ServiceList{}
	case NodeObjectType:
		return &NodeList{}
	default:
		panic(fmt.Sprintf("No ApiObjectType %v", ty))
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
		panic(fmt.Sprintf("No ApiObjectType %v", ty))
	}
	return nil
}

func GetApiObjectsURL(ty ApiObjectType) string {
	switch ty {
	case PodObjectType:
		return api.PodsURL
	case ServiceObjectType:
		return api.ServicesURL
	case NodeObjectType:
		return api.NodesURL
	default:
		panic(fmt.Sprintf("No ApiObjectType %v", ty))
	}
}

func GetWatchApiObjectsURL(ty ApiObjectType) string {
	switch ty {
	case PodObjectType:
		return api.WatchPodsURL
	case ServiceObjectType:
		return api.WatchServicesURL
	case NodeObjectType:
		return api.WatchNodesURL
	default:
		panic(fmt.Sprintf("No ApiObjectType %v", ty))
	}
}

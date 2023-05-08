package core

import (
	"fmt"
	"minik8s/pkg/api"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/types"
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

	// CreateFromEtcdString is for create by unmarshal an
	// ApiObject from etcd storage value (stored as string type)
	CreateFromEtcdString(str string) error

	// GenerateOwnerReference is used to generate OwnerReference
	// for filling meta.ObjectMeta OwnerReference[] field of
	// other ApiObject owned by it
	GenerateOwnerReference() meta.OwnerReference

	// AppendOwnerReference append new OwnerReference to
	// meta.ObjectMeta OwnerReference[] field of ApiObject
	AppendOwnerReference(meta.OwnerReference)
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

func CreateApiObject(ty types.ApiObjectType) IApiObject {
	switch ty {
	case types.PodObjectType:
		return &Pod{}
	case types.ServiceObjectType:
		return &Service{}
	case types.NodeObjectType:
		return &Node{}
	case types.ReplicasetObjectType:
		return &ReplicaSet{}
	case types.ErrorObjectType:
		return &ErrorApiObject{}
	default:
		panic(fmt.Sprintf("No ApiObjectType %v", ty))
	}
	return nil
}

func CreateApiObjectList(ty types.ApiObjectType) IApiObjectList {
	switch ty {
	case types.PodObjectType:
		return &PodList{}
	case types.ServiceObjectType:
		return &ServiceList{}
	case types.NodeObjectType:
		return &NodeList{}
	case types.ReplicasetObjectType:
		return &ReplicaSetList{}
	default:
		panic(fmt.Sprintf("No ApiObjectType %v", ty))
	}
	return nil
}

func CreateApiObjectStatus(ty types.ApiObjectType) IApiObjectStatus {
	switch ty {
	case types.PodObjectType:
		return &PodStatus{}
	case types.ServiceObjectType:
		return &ServiceStatus{}
	case types.NodeObjectType:
		return &NodeStatus{}
	case types.ReplicasetObjectType:
		return &ReplicaSetStatus{}
	default:
		panic(fmt.Sprintf("No ApiObjectType %v", ty))
	}
	return nil
}

func GetApiObjectsURL(ty types.ApiObjectType) string {
	switch ty {
	case types.PodObjectType:
		return api.PodsURL
	case types.ServiceObjectType:
		return api.ServicesURL
	case types.NodeObjectType:
		return api.NodesURL
	default:
		panic(fmt.Sprintf("No ApiObjectType %v", ty))
	}
}

func GetWatchApiObjectsURL(ty types.ApiObjectType) string {
	switch ty {
	case types.PodObjectType:
		return api.WatchPodsURL
	case types.ServiceObjectType:
		return api.WatchServicesURL
	case types.NodeObjectType:
		return api.WatchNodesURL
	default:
		panic(fmt.Sprintf("No ApiObjectType %v", ty))
	}
}

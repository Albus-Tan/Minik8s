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

	// DeleteOwnerReference delete OwnerReference of uid from
	// meta.ObjectMeta OwnerReference[] field of ApiObject
	DeleteOwnerReference(uid types.UID)

	PrintBrief()
}

type IApiObjectList interface {
	JsonUnmarshal(data []byte) error
	JsonMarshal() ([]byte, error)
	AddItemFromStr(objectStr string) error
	AppendItemsFromStr(objectStrs []string) error
	GetItems() any
	GetIApiObjectArr() []IApiObject
	PrintBrief()
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
	case types.HorizontalPodAutoscalerObjectType:
		return &HorizontalPodAutoscaler{}
	case types.FuncTemplateObjectType:
		return &Func{}
	case types.ErrorObjectType:
		return &ErrorApiObject{}
	case types.JobObjectType:
		return &Job{}
	case types.HeartbeatObjectType:
		return &Heartbeat{}
	case types.DnsObjectType:
		return &DNS{}
	default:
		panic(fmt.Sprintf("No ApiObjectType %v", ty))
	}
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
	case types.HorizontalPodAutoscalerObjectType:
		return &HorizontalPodAutoscalerList{}
	case types.JobObjectType:
		return &JobList{}
	case types.HeartbeatObjectType:
		return &HeartbeatList{}
	case types.FuncTemplateObjectType:
		return &FuncList{}
	case types.DnsObjectType:
		return &DnsList{}
	default:
		panic(fmt.Sprintf("No ApiObjectType %v", ty))
	}
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
	case types.HorizontalPodAutoscalerObjectType:
		return &HorizontalPodAutoscalerStatus{}
	case types.FuncTemplateObjectType:
		return &FuncStatus{}
	case types.JobObjectType:
		return &JobStatus{}
	case types.HeartbeatObjectType:
		return &HeartbeatStatus{}
	case types.DnsObjectType:
		return &DnsStatus{}
	default:
		panic(fmt.Sprintf("No ApiObjectType %v", ty))
	}
}

func GetApiObjectsURL(ty types.ApiObjectType) string {
	switch ty {
	case types.PodObjectType:
		return api.PodsURL
	case types.ServiceObjectType:
		return api.ServicesURL
	case types.NodeObjectType:
		return api.NodesURL
	case types.ReplicasetObjectType:
		return api.ReplicaSetsURL
	case types.HorizontalPodAutoscalerObjectType:
		return api.HorizontalPodAutoscalersURL
	case types.FuncTemplateObjectType:
		return api.FuncTemplatesURL
	case types.JobObjectType:
		return api.JobsURL
	case types.HeartbeatObjectType:
		return api.HeartbeatsURL
	case types.DnsObjectType:
		return api.DNSsURL
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
	case types.ReplicasetObjectType:
		return api.WatchReplicaSetsURL
	case types.HorizontalPodAutoscalerObjectType:
		return api.WatchHorizontalPodAutoscalersURL
	case types.JobObjectType:
		return api.WatchJobsURL
	case types.HeartbeatObjectType:
		return api.WatchHeartbeatsURL
	case types.DnsObjectType:
		return api.WatchDNSsURL
	case types.FuncTemplateObjectType:
		return api.WatchFuncTemplatesURL
	default:
		panic(fmt.Sprintf("No ApiObjectType %v", ty))
	}
}

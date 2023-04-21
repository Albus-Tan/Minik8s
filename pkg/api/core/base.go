package core

import "minik8s/pkg/api/types"

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
}

func CreateApiObject(ty ApiObjectType) IApiObject {
	switch ty {
	case PodObjectType:
		return Pod{}
	case ServiceObjectType:
		return Service{}
	case NodeObjectType:
		return Node{}
	default:
	}
	return nil
}

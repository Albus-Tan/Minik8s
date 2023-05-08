package types

import "time"

type UID = string
type Time = time.Time

type ApiObjectType string

// These are the valid ApiObjectType.
// Kind field in TypeMeta should be one of these
const (
	ErrorObjectType      ApiObjectType = "Error"
	PodObjectType        ApiObjectType = "Pod"
	ServiceObjectType    ApiObjectType = "Service"
	ReplicasetObjectType ApiObjectType = "ReplicaSet"
	NodeObjectType       ApiObjectType = "Node"
)

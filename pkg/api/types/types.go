package types

import (
	"time"
)

type UID = string
type Time = time.Time

type ApiObjectType string

// These are the valid ApiObjectType.
// Kind field in TypeMeta should be one of these
const (
	ErrorObjectType                   ApiObjectType = "Error"
	PodObjectType                     ApiObjectType = "Pod"
	ServiceObjectType                 ApiObjectType = "Service"
	ReplicasetObjectType              ApiObjectType = "ReplicaSet"
	HorizontalPodAutoscalerObjectType ApiObjectType = "HorizontalPodAutoscaler"
	NodeObjectType                    ApiObjectType = "Node"
	JobObjectType                     ApiObjectType = "Job"
	HeartbeatObjectType               ApiObjectType = "Heartbeat"
	FuncTemplateObjectType            ApiObjectType = "Func"
	DnsObjectType                     ApiObjectType = "DNS"
)

// ResourceName is the name identifying various resources in a ResourceList.
type ResourceName string

// Resource names must be not more than 63 characters, consisting of upper- or lower-case alphanumeric characters,
// with the -, _, and . characters allowed anywhere, except the first or last character.
// The default convention, matching that for annotations, is to use lower-case names, with dashes, rather than
// camel case, separating compound words.
// Fully-qualified resource typenames are constructed from a DNS-style subdomain, followed by a slash `/` and a name.
const (
	// ResourceCPU CPU, in cores. (500m = .5 cores)
	ResourceCPU ResourceName = "cpu"
	// ResourceMemory Memory, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	ResourceMemory ResourceName = "memory"
	// ResourceStorage Volume size, in bytes (e,g. 5Gi = 5GiB = 5 * 1024 * 1024 * 1024)
	ResourceStorage ResourceName = "storage"
	// ResourceEphemeralStorage Local ephemeral storage, in bytes. (500Gi = 500GiB = 500 * 1024 * 1024 * 1024)
	// The resource name for ResourceEphemeralStorage is alpha and it can change across releases.
	ResourceEphemeralStorage ResourceName = "ephemeral-storage"
)

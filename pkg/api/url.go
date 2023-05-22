package api

const StatusSuffix = "/status"

// Clear all

const ClearAllURL = "/clear"

// ------------------ REST API ---------------------
// Pod
const (
	PodsURL                = "/api/pods/"
	PodURL                 = "/api/pods/:name"
	WatchPodsURL           = "/api/watch/pods/"
	WatchPodURL            = "/api/watch/pods/:name"
	PodStatusURL           = "/api/pods/:name/status"
	PodsOnSpecifiedNodeURL = "/api/pods/nodes/:node"
)

// Node
const (
	NodesURL      = "/api/nodes/"
	NodeURL       = "/api/nodes/:name"
	WatchNodesURL = "/api/watch/nodes"
	WatchNodeURL  = "/api/watch/nodes/:name"
	NodeStatusURL = "/api/nodes/:name/status"
)

// Service
const (
	ServicesURL      = "/api/services/"
	ServiceURL       = "/api/services/:name"
	WatchServicesURL = "/api/watch/services/"
	WatchServiceURL  = "/api/watch/services/:name"
	ServiceStatusURL = "/api/services/:name/status"
)

// ReplicaSet
const (
	ReplicaSetsURL      = "/api/replicasets/"
	ReplicaSetURL       = "/api/replicasets/:name"
	WatchReplicaSetsURL = "/api/watch/replicasets/"
	WatchReplicaSetURL  = "/api/watch/replicasets/:name"
	ReplicaSetStatusURL = "/api/replicasets/:name/status"
)

// HorizontalPodAutoscaler
const (
	HorizontalPodAutoscalersURL      = "/api/hpa/"
	HorizontalPodAutoscalerURL       = "/api/hpa/:name"
	WatchHorizontalPodAutoscalersURL = "/api/watch/hpa/"
	WatchHorizontalPodAutoscalerURL  = "/api/watch/hpa/:name"
	HorizontalPodAutoscalerStatusURL = "/api/hpa/:name/status"
)

// Job
const (
	JobsURL      = "/api/jobs/"
	JobURL       = "/api/jobs/:name"
	WatchJobsURL = "/api/watch/jobs"
	WatchJobURL  = "/api/watch/jobs/:name"
	JobStatusURL = "/api/jobs/:name/status"
)

// Heartbeat
const (
	HeartbeatsURL      = "/api/heartbeats/"
	HeartbeatURL       = "/api/heartbeats/:name"
	WatchHeartbeatsURL = "/api/watch/heartbeats"
	WatchHeartbeatURL  = "/api/watch/heartbeats/:name"
	HeartbeatStatusURL = "/api/heartbeats/:name/status"
)

// ------------------ Test API ---------------------

const (
	TestsURL = "/api/tests/"
	TestURL  = "/api/tests/:name"
)

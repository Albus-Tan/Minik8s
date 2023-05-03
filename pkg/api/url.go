package api

const StatusSuffix = "/status"

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

// ------------------ Test API ---------------------

const (
	TestsURL = "/api/tests/"
	TestURL  = "/api/tests/:name"
)

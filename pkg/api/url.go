package api

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

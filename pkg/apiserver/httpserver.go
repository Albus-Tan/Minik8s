package apiserver

import (
	"github.com/gin-gonic/gin"
	"minik8s/pkg/api"
	"minik8s/pkg/apiserver/handlers"
)

type HttpServer interface {
	Run(addr string) (err error)
	BindHandlers()
}

func NewHttpServer() HttpServer {
	return &httpServer{
		router: gin.Default(),
	}
}

type httpServer struct {
	router *gin.Engine
}

func (h httpServer) Run(addr string) (err error) {
	return h.router.Run(addr)
}

func (h httpServer) BindHandlers() {

	// Clear all
	h.router.GET(api.ClearAllURL, handlers.HandleClearAll)

	/*--------------------- Pod ---------------------*/
	// Create a Pod
	// POST /api/pods/
	h.router.POST(api.PodsURL, handlers.HandlePostPod)
	// Update/Replace the specified Pod
	// PUT /api/pods/{name}
	h.router.PUT(api.PodURL, handlers.HandlePutPod)
	// Delete a Pod
	// DELETE /api/pods/{name}
	h.router.DELETE(api.PodURL, handlers.HandleDeletePod)
	// Read the specified Pod
	// GET /api/pods/{name}
	h.router.GET(api.PodURL, handlers.HandleGetPod)
	// List or watch objects of kind Pod
	// GET /api/pods
	h.router.GET(api.PodsURL, handlers.HandleGetPods)
	// Watch changes to an object of kind Pod
	// GET /api/watch/pods/{name}
	h.router.GET(api.WatchPodURL, handlers.HandleWatchPod)
	// Watch individual changes to a list of Pod
	// GET /api/watch/pods
	h.router.GET(api.WatchPodsURL, handlers.HandleWatchPods)
	/*--------------------- Pod Status ---------------------*/
	// Read status of the specified Pod
	// GET /api/pods/{name}/status
	h.router.GET(api.PodStatusURL, handlers.HandleGetPodStatus)
	// Replace status of the specified Pod
	// PUT /api/pods/{name}/status
	h.router.PUT(api.PodStatusURL, handlers.HandlePutPodStatus)

	/*--------------------- Node ---------------------*/
	// Create a Node
	// POST /api/nodes
	h.router.POST(api.NodesURL, handlers.HandlePostNode)
	// Update/Replace the specified Node
	// PUT /api/nodes/{name}
	h.router.PUT(api.NodeURL, handlers.HandlePutNode)
	// Delete a Node
	// DELETE /api/nodes/{name}
	h.router.DELETE(api.NodeURL, handlers.HandleDeleteNode)
	// Delete all Nodes
	// DELETE /api/nodes
	h.router.DELETE(api.NodesURL, handlers.HandleDeleteNodes)
	// Read the specified Node
	// GET /api/nodes/{name}
	h.router.GET(api.NodeURL, handlers.HandleGetNode)
	// List or watch objects of kind Node
	// GET /api/nodes
	h.router.GET(api.NodesURL, handlers.HandleGetNodes)
	// Watch changes to an object of kind Node
	// GET /api/watch/nodes/{name}
	h.router.GET(api.WatchNodeURL, handlers.HandleWatchNode)
	// Watch individual changes to a list of Node
	// GET /api/watch/nodes
	h.router.GET(api.WatchNodesURL, handlers.HandleWatchNodes)
	/*--------------------- Node Status ---------------------*/
	// Read status of the specified Node
	// GET /api/nodes/{name}/status
	h.router.GET(api.NodeStatusURL, handlers.HandleGetNodeStatus)
	// Update/Replace status of the specified Node
	// PUT /api/nodes/{name}/status
	h.router.PUT(api.NodeStatusURL, handlers.HandlePutNodeStatus)

	/*--------------------- Service ---------------------*/
	// Create a Service
	// POST /api/services
	h.router.POST(api.ServicesURL, handlers.HandlePostService)
	// Update/Replace the specified Service
	// PUT /api/services/{name}
	h.router.PUT(api.ServiceURL, handlers.HandlePutService)
	// Delete a Service
	// DELETE /api/services/{name}
	h.router.DELETE(api.ServiceURL, handlers.HandleDeleteService)
	// Read the specified Service
	// GET /api/services/{name}
	h.router.GET(api.ServiceURL, handlers.HandleGetService)
	// List or watch objects of kind Service
	// GET /api/services
	h.router.GET(api.ServicesURL, handlers.HandleGetServices)
	// Watch changes to an object of kind Service
	// GET /api/watch/services/{name}
	h.router.GET(api.WatchServiceURL, handlers.HandleWatchService)
	// Watch individual changes to a list of Service
	// GET /api/watch/services
	h.router.GET(api.WatchServicesURL, handlers.HandleWatchServices)
	/*--------------------- Service Status ---------------------*/
	// Read status of the specified Service
	// GET /api/services/{name}/status
	h.router.GET(api.ServiceStatusURL, handlers.HandleGetServiceStatus)
	// Replace status of the specified Service
	// PUT /api/services/{name}/status
	h.router.PUT(api.ServiceStatusURL, handlers.HandlePutServiceStatus)

	/*--------------------- ReplicaSet ---------------------*/
	// Create a ReplicaSet
	// POST /api/replicasets
	h.router.POST(api.ReplicaSetsURL, handlers.HandlePostReplicaSet)
	// Update/Replace the specified ReplicaSet
	// PUT /api/replicasets/{name}
	h.router.PUT(api.ReplicaSetURL, handlers.HandlePutReplicaSet)
	// Delete a ReplicaSet
	// DELETE /api/replicasets/{name}
	h.router.DELETE(api.ReplicaSetURL, handlers.HandleDeleteReplicaSet)
	// Read the specified ReplicaSet
	// GET /api/replicasets/{name}
	h.router.GET(api.ReplicaSetURL, handlers.HandleGetReplicaSet)
	// List or watch objects of kind ReplicaSet
	// GET /api/replicasets
	h.router.GET(api.ReplicaSetsURL, handlers.HandleGetReplicaSets)
	// Watch changes to an object of kind ReplicaSet
	// GET /api/watch/replicasets/{name}
	h.router.GET(api.WatchReplicaSetURL, handlers.HandleWatchReplicaSet)
	// Watch individual changes to a list of ReplicaSet
	// GET /api/watch/replicasets
	h.router.GET(api.WatchReplicaSetsURL, handlers.HandleWatchReplicaSets)
	/*--------------------- ReplicaSet Status ---------------------*/
	// Read status of the specified ReplicaSet
	// GET /api/replicasets/{name}/status
	h.router.GET(api.ReplicaSetStatusURL, handlers.HandleGetReplicaSetStatus)
	// Replace status of the specified ReplicaSet
	// PUT /api/replicasets/{name}/status
	h.router.PUT(api.ReplicaSetStatusURL, handlers.HandlePutReplicaSetStatus)

	/*--------------------- HorizontalPodAutoscaler ---------------------*/
	// Create a HorizontalPodAutoscaler
	// POST /api/hpa
	h.router.POST(api.HorizontalPodAutoscalersURL, handlers.HandlePostHorizontalPodAutoscaler)
	// Update/Replace the specified HorizontalPodAutoscaler
	// PUT /api/hpa/{name}
	h.router.PUT(api.HorizontalPodAutoscalerURL, handlers.HandlePutHorizontalPodAutoscaler)
	// Delete a HorizontalPodAutoscaler
	// DELETE /api/hpa/{name}
	h.router.DELETE(api.HorizontalPodAutoscalerURL, handlers.HandleDeleteHorizontalPodAutoscaler)
	// Read the specified HorizontalPodAutoscaler
	// GET /api/hpa/{name}
	h.router.GET(api.HorizontalPodAutoscalerURL, handlers.HandleGetHorizontalPodAutoscaler)
	// List or watch objects of kind HorizontalPodAutoscaler
	// GET /api/hpa
	h.router.GET(api.HorizontalPodAutoscalersURL, handlers.HandleGetHorizontalPodAutoscalers)
	// Watch changes to an object of kind HorizontalPodAutoscaler
	// GET /api/watch/hpa/{name}
	h.router.GET(api.WatchHorizontalPodAutoscalerURL, handlers.HandleWatchHorizontalPodAutoscaler)
	// Watch individual changes to a list of HorizontalPodAutoscaler
	// GET /api/watch/hpa
	h.router.GET(api.WatchHorizontalPodAutoscalersURL, handlers.HandleWatchHorizontalPodAutoscalers)
	/*--------------------- HorizontalPodAutoscaler Status ---------------------*/
	// Read status of the specified HorizontalPodAutoscaler
	// GET /api/hpa/{name}/status
	h.router.GET(api.HorizontalPodAutoscalerStatusURL, handlers.HandleGetHorizontalPodAutoscalerStatus)
	// Replace status of the specified HorizontalPodAutoscaler
	// PUT /api/hpa/{name}/status
	h.router.PUT(api.HorizontalPodAutoscalerStatusURL, handlers.HandlePutHorizontalPodAutoscalerStatus)

	/*--------------------- Job ---------------------*/
	// Create a Job
	// POST /api/jobs
	h.router.POST(api.JobsURL, handlers.HandlePostJob)
	// Update/Replace the specified Job
	// PUT /api/jobs/{name}
	h.router.PUT(api.JobURL, handlers.HandlePutJob)
	// Delete a Job
	// DELETE /api/jobs/{name}
	h.router.DELETE(api.JobURL, handlers.HandleDeleteJob)
	// Read the specified Job
	// GET /api/jobs/{name}
	h.router.GET(api.JobURL, handlers.HandleGetJob)
	// List or watch objects of kind Job
	// GET /api/jobs
	h.router.GET(api.JobsURL, handlers.HandleGetJobs)
	// Watch changes to an object of kind Job
	// GET /api/watch/jobs/{name}
	h.router.GET(api.WatchJobURL, handlers.HandleWatchJob)
	// Watch individual changes to a list of Job
	// GET /api/watch/jobs
	h.router.GET(api.WatchJobsURL, handlers.HandleWatchJobs)
	/*--------------------- Job Status ---------------------*/
	// Read status of the specified Job
	// GET /api/jobs/{name}/status
	h.router.GET(api.JobStatusURL, handlers.HandleGetJobStatus)
	// Replace status of the specified Job
	// PUT /api/jobs/{name}/status
	h.router.PUT(api.JobStatusURL, handlers.HandlePutJobStatus)

	/*--------------------- Heartbeat ---------------------*/
	// Create a Heartbeat
	// POST /api/heartbeats
	h.router.POST(api.HeartbeatsURL, handlers.HandlePostHeartbeat)
	// Update/Replace the specified Heartbeat
	// PUT /api/heartbeats/{name}
	h.router.PUT(api.HeartbeatURL, handlers.HandlePutHeartbeat)
	// Delete a Heartbeat
	// DELETE /api/heartbeats/{name}
	h.router.DELETE(api.HeartbeatURL, handlers.HandleDeleteHeartbeat)
	// Read the specified Heartbeat
	// GET /api/heartbeats/{name}
	h.router.GET(api.HeartbeatURL, handlers.HandleGetHeartbeat)
	// List or watch objects of kind Heartbeat
	// GET /api/heartbeats
	h.router.GET(api.HeartbeatsURL, handlers.HandleGetHeartbeats)
	// Watch changes to an object of kind Heartbeat
	// GET /api/watch/heartbeats/{name}
	h.router.GET(api.WatchHeartbeatURL, handlers.HandleWatchHeartbeat)
	// Watch individual changes to a list of Heartbeat
	// GET /api/watch/heartbeats
	h.router.GET(api.WatchHeartbeatsURL, handlers.HandleWatchHeartbeats)
	/*--------------------- Heartbeat Status ---------------------*/
	// Read status of the specified Heartbeat
	// GET /api/heartbeats/{name}/status
	h.router.GET(api.HeartbeatStatusURL, handlers.HandleGetHeartbeatStatus)
	// Replace status of the specified Heartbeat
	// PUT /api/heartbeats/{name}/status
	h.router.PUT(api.HeartbeatStatusURL, handlers.HandlePutHeartbeatStatus)
}

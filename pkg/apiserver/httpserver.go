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
}

//func handleGetTest(c *gin.Context) {
//	c.JSON(200, gin.H{
//		"message": "hello world",
//	})
//}

package handlers

import (
	"github.com/gin-gonic/gin"
	"minik8s/pkg/api"
	"minik8s/pkg/api/types"
)

/*--------------------- Pod ---------------------*/
//	log.Printf(c.Request.URL.Path) // /api/pods/actual-name
//	log.Printf(c.FullPath())       // /api/pods/:name

func HandlePostPod(c *gin.Context) {
	handlePostObject(c, types.PodObjectType)
}

func HandlePutPod(c *gin.Context) {
	handlePutObject(c, types.PodObjectType)
}

func HandleDeletePod(c *gin.Context) {
	handleDeleteObject(c, types.PodObjectType)
}

func HandleGetPod(c *gin.Context) {
	handleGetObject(c, types.PodObjectType)
}

func HandleGetPods(c *gin.Context) {
	handleGetObjects(c, types.PodObjectType)
}

func HandleWatchPod(c *gin.Context) {
	resourceURL := api.PodsURL + c.Param("name")
	handleWatchObjectAndStatus(c, types.PodObjectType, resourceURL)
}

func HandleWatchPods(c *gin.Context) {
	resourceURL := api.PodsURL
	handleWatchObjectsAndStatus(c, types.PodObjectType, resourceURL)
}

func HandleGetPodStatus(c *gin.Context) {
	resourceURL := api.PodsURL + c.Param("name")
	handleGetObjectStatus(c, types.PodObjectType, resourceURL)
}

func HandlePutPodStatus(c *gin.Context) {
	etcdURL := api.PodsURL + c.Param("name")
	handlePutObjectStatus(c, types.PodObjectType, etcdURL)
}

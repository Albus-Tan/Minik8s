package handlers

import (
	"github.com/gin-gonic/gin"
	"minik8s/pkg/api"
	"minik8s/pkg/api/core"
)

/*--------------------- Pod ---------------------*/
//	log.Printf(c.Request.URL.Path) // /api/pods/actual-name
//	log.Printf(c.FullPath())       // /api/pods/:name

func HandlePostPod(c *gin.Context) {
	handlePostObject(c, core.PodObjectType)
}

func HandlePutPod(c *gin.Context) {
	handlePutObject(c, core.PodObjectType)
}

func HandleDeletePod(c *gin.Context) {
	handleDeleteObject(c, core.PodObjectType)
}

func HandleGetPod(c *gin.Context) {
	handleGetObject(c, core.PodObjectType)
}

func HandleGetPods(c *gin.Context) {
	handleGetObjects(c, core.PodObjectType)
}

func HandleWatchPod(c *gin.Context) {
	resourceURL := api.PodsURL + c.Param("name")
	handleWatchObject(c, core.PodObjectType, resourceURL)
}

func HandleWatchPods(c *gin.Context) {
	resourceURL := api.PodsURL
	handleWatchObjects(c, core.PodObjectType, resourceURL)
}

func HandleGetPodStatus(c *gin.Context) {
	resourceURL := api.PodsURL + c.Param("name")
	handleGetObjectStatus(c, core.PodObjectType, resourceURL)
}

func HandlePutPodStatus(c *gin.Context) {
	etcdURL := api.PodsURL + c.Param("name")
	handlePutObjectStatus(c, core.PodObjectType, etcdURL)
}

package handlers

import (
	"github.com/gin-gonic/gin"
	"minik8s/pkg/api"
	"minik8s/pkg/api/types"
)

/*--------------------- HorizontalPodAutoscaler---------------------*/

func HandlePostHorizontalPodAutoscaler(c *gin.Context) {
	handlePostObject(c, types.HorizontalPodAutoscalerObjectType)
}

func HandlePutHorizontalPodAutoscaler(c *gin.Context) {
	handlePutObject(c, types.HorizontalPodAutoscalerObjectType)
}

func HandleDeleteHorizontalPodAutoscaler(c *gin.Context) {
	handleDeleteObject(c, types.HorizontalPodAutoscalerObjectType)
}

func HandleGetHorizontalPodAutoscaler(c *gin.Context) {
	handleGetObject(c, types.HorizontalPodAutoscalerObjectType)
}

func HandleGetHorizontalPodAutoscalers(c *gin.Context) {
	handleGetObjects(c, types.HorizontalPodAutoscalerObjectType)
}

func HandleWatchHorizontalPodAutoscaler(c *gin.Context) {
	resourceURL := api.HorizontalPodAutoscalersURL + c.Param("name")
	handleWatchObjectAndStatus(c, types.HorizontalPodAutoscalerObjectType, resourceURL)
}

func HandleWatchHorizontalPodAutoscalers(c *gin.Context) {
	resourceURL := api.HorizontalPodAutoscalersURL
	handleWatchObjectsAndStatus(c, types.HorizontalPodAutoscalerObjectType, resourceURL)
}

func HandleGetHorizontalPodAutoscalerStatus(c *gin.Context) {
	resourceURL := api.HorizontalPodAutoscalersURL + c.Param("name")
	handleGetObjectStatus(c, types.HorizontalPodAutoscalerObjectType, resourceURL)
}

func HandlePutHorizontalPodAutoscalerStatus(c *gin.Context) {
	etcdURL := api.HorizontalPodAutoscalersURL + c.Param("name")
	handlePutObjectStatus(c, types.HorizontalPodAutoscalerObjectType, etcdURL)
}

package handlers

import (
	"github.com/gin-gonic/gin"
	"minik8s/pkg/api"
	"minik8s/pkg/api/types"
)

/*--------------------- Service ---------------------*/

func HandlePostService(c *gin.Context) {
	handlePostObject(c, types.ServiceObjectType)
}

func HandlePutService(c *gin.Context) {
	handlePutObject(c, types.ServiceObjectType)
}

func HandleDeleteService(c *gin.Context) {
	handleDeleteObject(c, types.ServiceObjectType)
}

func HandleGetService(c *gin.Context) {
	handleGetObject(c, types.ServiceObjectType)
}

func HandleGetServices(c *gin.Context) {
	handleGetObjects(c, types.ServiceObjectType)
}

func HandleWatchService(c *gin.Context) {
	resourceURL := api.ServicesURL + c.Param("name")
	handleWatchObjectAndStatus(c, types.ServiceObjectType, resourceURL)
}

func HandleWatchServices(c *gin.Context) {
	resourceURL := api.ServicesURL
	handleWatchObjectsAndStatus(c, types.ServiceObjectType, resourceURL)
}

func HandleGetServiceStatus(c *gin.Context) {
	resourceURL := api.ServicesURL + c.Param("name")
	handleGetObjectStatus(c, types.ServiceObjectType, resourceURL)
}

func HandlePutServiceStatus(c *gin.Context) {
	etcdURL := api.ServicesURL + c.Param("name")
	handlePutObjectStatus(c, types.ServiceObjectType, etcdURL)
}

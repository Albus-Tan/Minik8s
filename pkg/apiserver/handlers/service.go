package handlers

import (
	"github.com/gin-gonic/gin"
	"minik8s/pkg/api"
	"minik8s/pkg/api/core"
)

/*--------------------- Service ---------------------*/

func HandlePostService(c *gin.Context) {
	handlePostObject(c, core.ServiceObjectType)
}

func HandlePutService(c *gin.Context) {
	handlePutObject(c, core.ServiceObjectType)
}

func HandleDeleteService(c *gin.Context) {
	handleDeleteObject(c, core.ServiceObjectType)
}

func HandleGetService(c *gin.Context) {
	handleGetObject(c, core.ServiceObjectType)
}

func HandleGetServices(c *gin.Context) {
	handleGetObjects(c, core.ServiceObjectType)
}

func HandleWatchService(c *gin.Context) {
	resourceURL := api.ServicesURL + c.Param("name")
	handleWatchObjectAndStatus(c, core.ServiceObjectType, resourceURL)
}

func HandleWatchServices(c *gin.Context) {
	resourceURL := api.ServicesURL
	handleWatchObjectsAndStatus(c, core.ServiceObjectType, resourceURL)
}

func HandleGetServiceStatus(c *gin.Context) {
	resourceURL := api.ServicesURL + c.Param("name")
	handleGetObjectStatus(c, core.ServiceObjectType, resourceURL)
}

func HandlePutServiceStatus(c *gin.Context) {
	etcdURL := api.ServicesURL + c.Param("name")
	handlePutObjectStatus(c, core.ServiceObjectType, etcdURL)
}

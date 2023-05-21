package handlers

import (
	"github.com/gin-gonic/gin"
	"minik8s/pkg/api"
	"minik8s/pkg/api/types"
)

/*--------------------- Job---------------------*/

func HandlePostJob(c *gin.Context) {
	handlePostObject(c, types.JobObjectType)
}

func HandlePutJob(c *gin.Context) {
	handlePutObject(c, types.JobObjectType)
}

func HandleDeleteJob(c *gin.Context) {
	handleDeleteObject(c, types.JobObjectType)
}

func HandleGetJob(c *gin.Context) {
	handleGetObject(c, types.JobObjectType)
}

func HandleGetJobs(c *gin.Context) {
	handleGetObjects(c, types.JobObjectType)
}

func HandleWatchJob(c *gin.Context) {
	resourceURL := api.JobsURL + c.Param("name")
	handleWatchObjectAndStatus(c, types.JobObjectType, resourceURL)
}

func HandleWatchJobs(c *gin.Context) {
	resourceURL := api.JobsURL
	handleWatchObjectsAndStatus(c, types.JobObjectType, resourceURL)
}

func HandleGetJobStatus(c *gin.Context) {
	resourceURL := api.JobsURL + c.Param("name")
	handleGetObjectStatus(c, types.JobObjectType, resourceURL)
}

func HandlePutJobStatus(c *gin.Context) {
	etcdURL := api.JobsURL + c.Param("name")
	handlePutObjectStatus(c, types.JobObjectType, etcdURL)
}

package handlers

import (
	"github.com/gin-gonic/gin"
	"minik8s/pkg/api"
	"minik8s/pkg/api/types"
)

/*--------------------- Heartbeat---------------------*/

func HandlePostHeartbeat(c *gin.Context) {
	handlePostObject(c, types.HeartbeatObjectType)
}

func HandlePutHeartbeat(c *gin.Context) {
	handlePutObject(c, types.HeartbeatObjectType)
}

func HandleDeleteHeartbeat(c *gin.Context) {
	handleDeleteObject(c, types.HeartbeatObjectType)
}

func HandleGetHeartbeat(c *gin.Context) {
	handleGetObject(c, types.HeartbeatObjectType)
}

func HandleGetHeartbeats(c *gin.Context) {
	handleGetObjects(c, types.HeartbeatObjectType)
}

func HandleWatchHeartbeat(c *gin.Context) {
	resourceURL := api.HeartbeatsURL + c.Param("name")
	handleWatchObjectAndStatus(c, types.HeartbeatObjectType, resourceURL)
}

func HandleWatchHeartbeats(c *gin.Context) {
	resourceURL := api.HeartbeatsURL
	handleWatchObjectsAndStatus(c, types.HeartbeatObjectType, resourceURL)
}

func HandleGetHeartbeatStatus(c *gin.Context) {
	resourceURL := api.HeartbeatsURL + c.Param("name")
	handleGetObjectStatus(c, types.HeartbeatObjectType, resourceURL)
}

func HandlePutHeartbeatStatus(c *gin.Context) {
	etcdURL := api.HeartbeatsURL + c.Param("name")
	handlePutObjectStatus(c, types.HeartbeatObjectType, etcdURL)
}

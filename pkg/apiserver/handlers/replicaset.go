package handlers

import (
	"github.com/gin-gonic/gin"
	"minik8s/pkg/api"
	"minik8s/pkg/api/types"
)

/*--------------------- ReplicaSet ---------------------*/

func HandlePostReplicaSet(c *gin.Context) {
	handlePostObject(c, types.ReplicasetObjectType)
}

func HandlePutReplicaSet(c *gin.Context) {
	handlePutObject(c, types.ReplicasetObjectType)
}

func HandleDeleteReplicaSet(c *gin.Context) {
	handleDeleteObject(c, types.ReplicasetObjectType)
}

func HandleGetReplicaSet(c *gin.Context) {
	handleGetObject(c, types.ReplicasetObjectType)
}

func HandleGetReplicaSets(c *gin.Context) {
	handleGetObjects(c, types.ReplicasetObjectType)
}

func HandleWatchReplicaSet(c *gin.Context) {
	resourceURL := api.ReplicaSetsURL + c.Param("name")
	handleWatchObjectAndStatus(c, types.ReplicasetObjectType, resourceURL)
}

func HandleWatchReplicaSets(c *gin.Context) {
	resourceURL := api.ReplicaSetsURL
	handleWatchObjectsAndStatus(c, types.ReplicasetObjectType, resourceURL)
}

func HandleGetReplicaSetStatus(c *gin.Context) {
	resourceURL := api.ReplicaSetsURL + c.Param("name")
	handleGetObjectStatus(c, types.ReplicasetObjectType, resourceURL)
}

func HandlePutReplicaSetStatus(c *gin.Context) {
	etcdURL := api.ReplicaSetsURL + c.Param("name")
	handlePutObjectStatus(c, types.ReplicasetObjectType, etcdURL)
}

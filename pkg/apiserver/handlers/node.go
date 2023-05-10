package handlers

import (
	"github.com/gin-gonic/gin"
	"minik8s/pkg/api"
	"minik8s/pkg/api/types"
)

/*--------------------- Node ---------------------*/

func HandlePostNode(c *gin.Context) {
	handlePostObject(c, types.NodeObjectType)
}

func HandlePutNode(c *gin.Context) {
	handlePutObject(c, types.NodeObjectType)
}

func HandleDeleteNode(c *gin.Context) {
	handleDeleteObject(c, types.NodeObjectType)
}

func HandleDeleteNodes(c *gin.Context) {
	// TODO: implement me
	panic("implement me")
}

func HandleGetNode(c *gin.Context) {
	handleGetObject(c, types.NodeObjectType)
}

func HandleGetNodes(c *gin.Context) {
	handleGetObjects(c, types.NodeObjectType)
}

func HandleWatchNode(c *gin.Context) {
	resourceURL := api.NodesURL + c.Param("name")
	handleWatchObjectAndStatus(c, types.NodeObjectType, resourceURL)
}

func HandleWatchNodes(c *gin.Context) {
	resourceURL := api.NodesURL
	handleWatchObjectsAndStatus(c, types.NodeObjectType, resourceURL)
}

func HandleGetNodeStatus(c *gin.Context) {
	resourceURL := api.NodesURL + c.Param("name")
	handleGetObjectStatus(c, types.NodeObjectType, resourceURL)
}

func HandlePutNodeStatus(c *gin.Context) {
	etcdURL := api.NodesURL + c.Param("name")
	handlePutObjectStatus(c, types.NodeObjectType, etcdURL)
}

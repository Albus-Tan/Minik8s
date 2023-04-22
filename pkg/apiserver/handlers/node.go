package handlers

import (
	"github.com/gin-gonic/gin"
	"minik8s/pkg/api"
	"minik8s/pkg/api/core"
)

/*--------------------- Node ---------------------*/

func HandlePostNode(c *gin.Context) {
	handlePostObject(c, core.NodeObjectType)
}

func HandlePutNode(c *gin.Context) {
	handlePutObject(c, core.NodeObjectType)
}

func HandleDeleteNode(c *gin.Context) {
	handleDeleteObject(c, core.NodeObjectType)
}

func HandleDeleteNodes(c *gin.Context) {
	// TODO: implement me
	panic("implement me")
}

func HandleGetNode(c *gin.Context) {
	handleGetObject(c, core.NodeObjectType)
}

func HandleGetNodes(c *gin.Context) {
	handleGetObjects(c, core.NodeObjectType)
}

func HandleWatchNode(c *gin.Context) {
	resourceURL := api.NodesURL + c.Param("name")
	handleWatchObjectAndStatus(c, core.NodeObjectType, resourceURL)
}

func HandleWatchNodes(c *gin.Context) {
	resourceURL := api.NodesURL
	handleWatchObjectsAndStatus(c, core.NodeObjectType, resourceURL)
}

func HandleGetNodeStatus(c *gin.Context) {
	resourceURL := api.NodesURL + c.Param("name")
	handleGetObjectStatus(c, core.NodeObjectType, resourceURL)
}

func HandlePutNodeStatus(c *gin.Context) {
	etcdURL := api.NodesURL + c.Param("name")
	handlePutObjectStatus(c, core.NodeObjectType, etcdURL)
}

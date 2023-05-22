package handlers

import (
	"github.com/gin-gonic/gin"
	"minik8s/pkg/api"
	"minik8s/pkg/api/types"
)

/*--------------------- DNS ---------------------*/

func HandlePostDNS(c *gin.Context) {
	handlePostObject(c, types.DnsObjectType)
}

func HandlePutDNS(c *gin.Context) {
	handlePutObject(c, types.DnsObjectType)
}

func HandleDeleteDNS(c *gin.Context) {
	handleDeleteObject(c, types.DnsObjectType)
}

func HandleGetDNS(c *gin.Context) {
	handleGetObject(c, types.DnsObjectType)
}

func HandleGetDNSs(c *gin.Context) {
	handleGetObjects(c, types.DnsObjectType)
}

func HandleWatchDNS(c *gin.Context) {
	resourceURL := api.DNSsURL + c.Param("name")
	handleWatchObjectAndStatus(c, types.DnsObjectType, resourceURL)
}

func HandleWatchDNSs(c *gin.Context) {
	resourceURL := api.DNSsURL
	handleWatchObjectsAndStatus(c, types.DnsObjectType, resourceURL)
}

func HandleGetDNSStatus(c *gin.Context) {
	resourceURL := api.DNSsURL + c.Param("name")
	handleGetObjectStatus(c, types.DnsObjectType, resourceURL)
}

func HandlePutDNSStatus(c *gin.Context) {
	etcdURL := api.DNSsURL + c.Param("name")
	handlePutObjectStatus(c, types.DnsObjectType, etcdURL)
}

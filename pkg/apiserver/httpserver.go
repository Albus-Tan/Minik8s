package apiserver

import "github.com/gin-gonic/gin"

type HttpServer interface {
	Run(addr string) (err error)
	BindHandlers()
}

func NewHttpServer() HttpServer {
	return &httpServer{
		router: gin.Default(),
	}
}

type httpServer struct {
	router *gin.Engine
}

func (h httpServer) Run(addr string) (err error) {
	return h.router.Run(addr)
}

func (h httpServer) BindHandlers() {
	//TODO implement me

	h.router.GET("/test", handleGetTest)
}

func handleGetTest(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "hello world",
	})
}

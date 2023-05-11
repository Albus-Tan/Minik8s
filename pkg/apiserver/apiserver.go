package apiserver

import (
	"minik8s/config"
	"minik8s/pkg/apiserver/etcd"
	"minik8s/pkg/logger"
)

type ApiServer interface {
	Run()
}

func New() ApiServer {
	return &apiServer{
		httpServer: NewHttpServer(),
		logger:     logger.ApiServerLogger,
	}
}

type apiServer struct {
	httpServer HttpServer
	logger     logger.Logger
}

func (a apiServer) Run() {
	a.logger.Printf("[apiserver] apiserver start\n")

	// etcd
	etcd.Init()
	defer etcd.Close()

	a.httpServer.BindHandlers()

	// Listen and Server in 0.0.0.0:8080
	err := a.httpServer.Run(config.Port)
	if err != nil {
		a.logger.Printf("[apiserver] httpserver start FAILED\n")
		a.logger.Fatal(err)
	}
}

//func (a apiServer) etcdApiTest() {
//	a.logger.Printf("[apiserver] start etcdApiTest\n")
//
//	_ = etcdPut("123", "12314333eee")
//	res, _ := etcdGet("123")
//	a.logger.Printf("[apiserver] expected %v, actual %v\n", "12314333eee", res)
//	_ = etcdDelete("123")
//	res, _ = etcdGet("123")
//	a.logger.Printf("[apiserver] expected %v, actual %v\n", "", res)
//	//_ = etcdClear()
//}

//func (a apiServer) etcdCheckVersionPutTest() {
//	a.logger.Printf("[apiserver] start etcdCheckVersionPutTest\n")
//
//	// _ = etcd.Put("123444", "11111")
//	_, _ = etcd.CheckVersionPut("123444", "12314333eee", "201")
//	_, version, _ := etcd.GetWithVersion("123444")
//	_, _ = etcd.CheckVersionPut("123444", "123", version)
//	_, _, _ = etcd.GetWithVersion("123444")
//}

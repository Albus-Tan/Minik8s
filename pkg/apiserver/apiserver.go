package apiserver

import (
	"log"
	"minik8s/pkg/apiserver/etcd"
)

type ApiServer interface {
	Run()
}

func New() ApiServer {
	return &apiServer{
		httpServer: NewHttpServer(),
	}
}

type apiServer struct {
	httpServer HttpServer
}

func (a apiServer) Run() {
	log.Printf("[apiserver] apiserver start\n")

	// etcd
	etcd.Init()
	defer etcd.Close()

	a.httpServer.BindHandlers()

	// Listen and Server in 0.0.0.0:8080
	err := a.httpServer.Run(":8080")
	if err != nil {
		log.Printf("[apiserver] httpserver start FAILED\n")
		log.Fatal(err)
	}
}

//func (a apiServer) etcdApiTest() {
//	log.Printf("[apiserver] start etcdApiTest\n")
//
//	_ = etcdPut("123", "12314333eee")
//	res, _ := etcdGet("123")
//	log.Printf("[apiserver] expected %v, actual %v\n", "12314333eee", res)
//	_ = etcdDelete("123")
//	res, _ = etcdGet("123")
//	log.Printf("[apiserver] expected %v, actual %v\n", "", res)
//	//_ = etcdClear()
//}

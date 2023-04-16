package apiserver

import (
	"log"
)

type ApiServer interface {
	Run()
}

func New() ApiServer {
	return &apiServer{}
}

type apiServer struct {
	name string
}

func (a apiServer) Run() {
	log.Printf("[apiserver] start function Run\n")
	initEtcd()

}

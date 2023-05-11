package main

import (
	"minik8s/pkg/apiserver"
	"minik8s/pkg/controller"
)

func main() {
	apiServer := apiserver.New()
	apiServer.Run()
	controllerManager := controller.NewControllerManager()
	controllerManager.Run()
}

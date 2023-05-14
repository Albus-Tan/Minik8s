package main

import (
	"minik8s/pkg/apiserver"
	"minik8s/pkg/controller"
	"minik8s/pkg/node"
	"minik8s/pkg/scheduler"
)

func main() {

	n := node.CreateMasterNode()
	defer node.DeleteNode(n)

	apiServer := apiserver.New()
	apiServer.Run()
	s := scheduler.NewScheduler()
	s.Run()
	controllerManager := controller.NewControllerManager()
	controllerManager.Run()
}

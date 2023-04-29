package main

import (
	"minik8s/pkg/controller"
)

func main() {
	controllerManager := controller.NewControllerManager()
	controllerManager.Run()
}

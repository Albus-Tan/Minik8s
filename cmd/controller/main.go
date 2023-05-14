package main

import (
	"context"
	"minik8s/pkg/controller"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	controllerManager := controller.NewControllerManager()
	controllerManager.Run(ctx, cancel)

	<-ctx.Done()
}

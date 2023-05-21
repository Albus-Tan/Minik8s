package main

import (
	"context"
	"log"
	"minik8s/pkg/apiserver"
	"minik8s/pkg/controller"
	"minik8s/pkg/gpu"
	"minik8s/pkg/node"
	"minik8s/pkg/node/heartbeat"
	"minik8s/pkg/scheduler"
	"time"
)

func main() {

	log.Printf("[Master] master start\n")
	defer log.Printf("[Master] master finish\n")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	apiServer := apiserver.New()
	apiServer.Run(cancel)
	time.Sleep(1 * time.Second)
	log.Printf("[Master] master apiServer running\n")

	n := node.CreateMasterNode()
	defer node.DeleteNode(n)

	log.Printf("[Master] master node running\n")

	s := scheduler.NewScheduler()
	s.Run(ctx, cancel)

	log.Printf("[Master] master scheduler running\n")

	controllerManager := controller.NewControllerManager()
	controllerManager.Run(ctx, cancel)

	log.Printf("[Master] master controllerManager running\n")

	gpuServer := gpu.NewServer()
	gpuServer.Run(ctx, cancel)

	log.Printf("[Master] master gpu server running\n")

	heartbeatWatcher := heartbeat.NewWatcher()
	heartbeatWatcher.Run(ctx, cancel)

	log.Printf("[Master] master heartbeat watcher running\n")

	log.Printf("[Master] master init finish\n")

	<-ctx.Done()
}

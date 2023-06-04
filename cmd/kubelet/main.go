package main

import (
	"context"
	"log"
	"minik8s/pkg/kubelet"
	"minik8s/pkg/node"
	"minik8s/pkg/node/heartbeat"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	n := node.CreateWorkerNode()
	defer node.DeleteNode(n)

	heartbeatSender := heartbeat.NewSender(n.UID)
	heartbeatSender.Run(ctx, cancel)

	k, err := kubelet.New(n)
	defer k.Close()
	if err != nil {
		log.Fatal(err.Error())
	}
	k.Run()

	<-ctx.Done()
}

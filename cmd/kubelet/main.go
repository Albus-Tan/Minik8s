package main

import (
	"log"
	"minik8s/pkg/kubelet"
	"minik8s/pkg/node"
)

func main() {
	n := node.CreateWorkerNode()
	k, err := kubelet.New(n)
	defer k.Close()
	if err != nil {
		log.Fatal(err.Error())
	}
	k.Run()
}

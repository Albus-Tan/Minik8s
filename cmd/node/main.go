package main

import (
	"minik8s/pkg/kubelet"
	"minik8s/pkg/node"
)

func main() {
	n := node.CreateWorkerNode()
	k := kubelet.New(n)
	k.Run()
}

package main

import (
	"minik8s/pkg/kubelet"
)

func main() {
	k := kubelet.New()
	k.Run()
}

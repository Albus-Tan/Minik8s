package main

import "minik8s/pkg/kubeproxy"

func main() {
	k := kubeproxy.New()
	k.Run()
}

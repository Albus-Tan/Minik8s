package main

import (
	"log"
	"minik8s/pkg/kubelet"
)

func main() {
	k, err := kubelet.New()
	defer k.Close()
	if err != nil {
		log.Fatal(err.Error())
	}
	k.Run()
}

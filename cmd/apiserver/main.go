package main

import "minik8s/pkg/apiserver"

func main() {
	apiServer := apiserver.New()
	apiServer.Run()
}

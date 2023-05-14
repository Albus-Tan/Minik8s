package main

import (
	"context"
	"minik8s/pkg/apiserver"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	apiServer := apiserver.New()
	apiServer.Run(cancel)

	<-ctx.Done()
}

package main

import (
	"context"
	"minik8s/pkg/gpu"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	gpuServer := gpu.NewServer()
	gpuServer.Run(ctx, cancel)

	<-ctx.Done()
}

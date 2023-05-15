package main

import (
	"context"
	"minik8s/pkg/scheduler"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := scheduler.NewScheduler()
	s.Run(ctx, cancel)

	<-ctx.Done()

}

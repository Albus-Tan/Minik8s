package main

import (
	"minik8s/pkg/scheduler"
)

func main() {
	s := scheduler.NewScheduler()
	s.Run()
}

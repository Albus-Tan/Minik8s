package main

import (
	"log"

	"github.com/moby/ipvs"
)

func main() {
	handle, err := ipvs.New("")
	if err != nil {
		log.Fatalf("ipvs.New: %s", err)
	}
	svcs, err := handle.GetServices()
	if err != nil {
		log.Fatalf("handle.GetServices: %s", err)
	}
	for _, svc := range svcs {
		println(svc.Port)
	}
}

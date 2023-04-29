package replicaset

import "log"

type Controller interface {
	Run()
}

func NewController() Controller {
	return &controller{}
}

type controller struct {
}

func (c *controller) Run() {
	//TODO implement me
	log.SetPrefix("[ReplicaSetController] ")
	log.Printf("start\n")
}

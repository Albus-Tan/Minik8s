package controller

import (
	"log"
	"minik8s/pkg/controller/replicaset"
)

type Manager interface {
	Run()
}

func NewControllerManager() Manager {
	return &manager{
		replicaSetController: replicaset.NewController(),
	}
}

type manager struct {
	replicaSetController replicaset.Controller
}

func (m *manager) Run() {
	//TODO implement me
	log.SetPrefix("[ControllerManager] ")
	log.Printf("manager start\n")

	m.replicaSetController.Run()

	panic("implement me")
}

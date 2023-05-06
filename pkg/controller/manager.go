package controller

import (
	"log"
	"minik8s/pkg/api/core"
	"minik8s/pkg/apiclient"
	client "minik8s/pkg/apiclient/interface"
	"minik8s/pkg/apiclient/listwatch"
	"minik8s/pkg/controller/cache"
	"minik8s/pkg/controller/replicaset"
)

type Manager interface {
	Run()
}

func NewControllerManager() Manager {

	// Client and Informer can be reused for same resource type
	podClient, podInformer := NewDefaultClientSet(core.PodObjectType)
	rsClient, rsInformer := NewDefaultClientSet(core.ReplicasetObjectType)

	return &manager{
		// Client
		podClient: podClient,
		rsClient:  rsClient,
		// Informer
		podInformer: podInformer,
		rsInformer:  rsInformer,
		// Controller
		replicaSetController: replicaset.NewReplicaSetController(podInformer, podClient, rsInformer, rsClient),
	}
}

type manager struct {
	// Client
	podClient client.Interface
	rsClient  client.Interface

	// Informer
	podInformer cache.Informer
	rsInformer  cache.Informer

	// Controller
	replicaSetController replicaset.ReplicaSetController
}

func NewDefaultClientSet(objType core.ApiObjectType) (client.Interface, cache.Informer) {
	restClient, _ := apiclient.NewRESTClient(objType)
	lw := listwatch.NewListWatchFromClient(restClient)
	informer := cache.NewDefaultInformer(lw, objType)
	return restClient, informer
}

func (m *manager) Run() {
	//TODO implement me
	log.SetPrefix("[ControllerManager] ")
	log.Printf("manager start\n")

	stopCh := make(chan struct{})

	// Run Informer
	m.podInformer.Run(stopCh)
	m.rsInformer.Run(stopCh)

	// Run Controller
	m.replicaSetController.Run()

	panic("implement me")

	// Stop controller manager and all related go routines
	// close(stopCh)
}

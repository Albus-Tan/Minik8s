package controller

import (
	"context"
	"log"
	"minik8s/pkg/api/types"
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
	podClient, podInformer := NewDefaultClientSet(types.PodObjectType)
	rsClient, rsInformer := NewDefaultClientSet(types.ReplicasetObjectType)

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

func NewDefaultClientSet(objType types.ApiObjectType) (client.Interface, cache.Informer) {
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
	ctx, cancel := context.WithCancel(context.Background())

	// Stop controller manager and all related go routines
	defer close(stopCh)
	defer cancel()

	// Run Informer
	m.podInformer.Run(stopCh)
	m.rsInformer.Run(stopCh)

	// Run Controller
	m.replicaSetController.Run(ctx)

}

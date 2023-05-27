package controller

import (
	"context"
	"minik8s/pkg/api/types"
	"minik8s/pkg/apiclient"
	client "minik8s/pkg/apiclient/interface"
	"minik8s/pkg/apiclient/listwatch"
	"minik8s/pkg/controller/cache"
	"minik8s/pkg/controller/dns"
	"minik8s/pkg/controller/podautoscaler"
	"minik8s/pkg/controller/replicaset"
	"minik8s/pkg/controller/serverless"
	"minik8s/pkg/logger"
)

type Manager interface {
	Run(ctx context.Context, cancel context.CancelFunc)
}

func NewControllerManager() Manager {

	// Client and Informer can be reused for same resource type
	podClient, podInformer := NewDefaultClientSet(types.PodObjectType)
	rsClient, rsInformer := NewDefaultClientSet(types.ReplicasetObjectType)
	hpaClient, hpaInformer := NewDefaultClientSet(types.HorizontalPodAutoscalerObjectType)
	dnsClient, dnsInformer := NewDefaultClientSet(types.DnsObjectType)
	funcTemplateClient, funcTemplateInformer := NewDefaultClientSet(types.FuncTemplateObjectType)
	serviceClient, _ := apiclient.NewRESTClient(types.ServiceObjectType)

	return &manager{
		// Client
		podClient:          podClient,
		rsClient:           rsClient,
		hpaClient:          hpaClient,
		serviceClient:      serviceClient,
		dnsClient:          dnsClient,
		funcTemplateClient: funcTemplateClient,
		// Informer
		podInformer:          podInformer,
		rsInformer:           rsInformer,
		hpaInformer:          hpaInformer,
		dnsInformer:          dnsInformer,
		funcTemplateInformer: funcTemplateInformer,
		// Controller
		replicaSetController: replicaset.NewReplicaSetController(podInformer, podClient, rsInformer, rsClient),
		horizontalController: podautoscaler.NewHorizontalController(podInformer, podClient, hpaInformer, hpaClient, rsInformer, rsClient),
		dnsController:        dns.NewDnsController(podClient, serviceClient, dnsInformer, dnsClient),
		serverlessController: serverless.NewServerlessController(funcTemplateInformer, funcTemplateClient, rsClient, serviceClient, podClient),
	}
}

type manager struct {
	// Client
	podClient          client.Interface
	rsClient           client.Interface
	hpaClient          client.Interface
	serviceClient      client.Interface
	dnsClient          client.Interface
	funcTemplateClient client.Interface
	// Informer
	podInformer          cache.Informer
	rsInformer           cache.Informer
	hpaInformer          cache.Informer
	dnsInformer          cache.Informer
	funcTemplateInformer cache.Informer
	// Controller
	replicaSetController replicaset.ReplicaSetController
	horizontalController podautoscaler.HorizontalController
	dnsController        dns.DnsController
	serverlessController serverless.ServerlessController
}

func NewDefaultClientSet(objType types.ApiObjectType) (client.Interface, cache.Informer) {
	restClient, _ := apiclient.NewRESTClient(objType)
	lw := listwatch.NewListWatchFromClient(restClient)
	informer := cache.NewDefaultInformer(lw, objType)
	return restClient, informer
}

func (m *manager) Run(ctx context.Context, cancel context.CancelFunc) {

	logger.ControllerManagerLogger.Printf("[ControllerManager] manager start\n")
	defer logger.ControllerManagerLogger.Printf("[ControllerManager] manager init finish\n")

	// Stop controller manager and all related go routines
	// use stopCh := ctx.Done()

	// Run Informer
	m.podInformer.Run(ctx.Done())
	m.rsInformer.Run(ctx.Done())
	m.hpaInformer.Run(ctx.Done())
	m.dnsInformer.Run(ctx.Done())
	m.funcTemplateInformer.Run(ctx.Done())

	// Run Controller
	m.replicaSetController.Run(ctx)
	m.horizontalController.Run(ctx)
	m.dnsController.Run(ctx)
	m.serverlessController.Run(ctx)
}

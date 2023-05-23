package dns

import (
	"context"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/types"
	client "minik8s/pkg/apiclient/interface"
	"minik8s/pkg/controller/cache"
	"minik8s/pkg/logger"
	"time"
)

type DnsController interface {
	Run(ctx context.Context)
}

func NewDnsController(podClient client.Interface, serviceClient client.Interface, DnsInformer cache.Informer, DnsClient client.Interface) DnsController {

	dnsc := &dnsController{
		Kind:          string(types.DnsObjectType),
		PodClient:     podClient,
		ServiceClient: serviceClient,
		DnsInformer:   DnsInformer,
		DnsClient:     DnsClient,
		queue:         cache.NewWorkQueue(),
	}

	_ = dnsc.DnsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    dnsc.addDNS,
		DeleteFunc: dnsc.deleteDNS,
	})

	return dnsc
}

type dnsController struct {
	Kind string

	DnsInformer   cache.Informer
	PodClient     client.Interface
	ServiceClient client.Interface
	DnsClient     client.Interface
	queue         cache.WorkQueue
}

func (dnsc *dnsController) Run(ctx context.Context) {

	go func() {
		logger.DNSControllerLogger.Printf("[DnsController] start\n")
		defer logger.DNSControllerLogger.Printf("[DnsController] finish\n")

		dnsc.runWorker(ctx)

		// wait for controller manager stop
		<-ctx.Done()
	}()
	return
}

func (dnsc *dnsController) DNSKeyFunc(DNS *core.DNS) string {
	return DNS.GetUID()
}

func (dnsc *dnsController) enqueueDNS(DNS *core.DNS) {
	key := dnsc.DNSKeyFunc(DNS)
	dnsc.queue.Enqueue(DNS)
	logger.DNSControllerLogger.Printf("enqueueDNS uid %s\n", key)
}

func (dnsc *dnsController) addDNS(obj interface{}) {
	DNS := obj.(*core.DNS)
	logger.DNSControllerLogger.Printf("Adding %s %s/%s\n", dnsc.Kind, DNS.Namespace, DNS.Name)
	dnsc.enqueueDNS(DNS)
}

func (dnsc *dnsController) deleteDNS(obj interface{}) {
	DNS := obj.(*core.DNS)

	logger.DNSControllerLogger.Printf("Deleting %s, uid %s\n", dnsc.Kind, DNS.UID)

	// TODO directly delete corresponding apiobject

}

const defaultWorkeDNSleepInterval = time.Duration(3) * time.Second

func (dnsc *dnsController) runWorker(ctx context.Context) {
	// go wait.UntilWithContext(ctx, dnsc.worker, time.Second)
	for {
		select {
		case <-ctx.Done():
			logger.DNSControllerLogger.Printf("[worker] ctx.Done() received, worker of DnsController exit\n")
			return
		default:
			for dnsc.processNextWorkItem() {
			}
			time.Sleep(defaultWorkeDNSleepInterval)
		}
	}
}

func (dnsc *dnsController) processNextWorkItem() bool {

	item, ok := dnsc.queue.Dequeue()
	if !ok {
		return false
	}

	dns := item.(*core.DNS)

	err := dnsc.processDNSCreate(dns)
	if err != nil {
		logger.DNSControllerLogger.Printf("[processDNSCreate] err: %v\n", err)
		// enqueue if error happen when processing
		dnsc.queue.Enqueue(dns)
		return false
	}

	return true
}

func (dnsc *dnsController) processDNSCreate(dns *core.DNS) error {

	// TODO
	// 	process dns create event

	return nil
}

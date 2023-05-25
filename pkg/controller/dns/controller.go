package dns

import (
	"context"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/generate"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/types"
	client "minik8s/pkg/apiclient/interface"
	"minik8s/pkg/controller/cache"
	"minik8s/pkg/logger"
	"strings"
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

	_, _, err := dnsc.ServiceClient.Delete(DNS.Status.ServiceUID)
	if err != nil {
		return
	}
	_, _, err = dnsc.PodClient.Delete(DNS.Status.PodUID)
	if err != nil {
		return
	}

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

	c := make([]string, 0)
	for _, m := range dns.Spec.Mappings {
		c = append(c, m.Path+"#"+m.Address)
	}

	pod := generate.EmptyPod()
	pod.Name = "dns-" + dns.UID
	pod.Spec = core.PodSpec{
		Containers: []core.Container{
			{
				Name:  "gateway-" + dns.UID,
				Image: "lwsg/gateway-runner:0.4",
				Env: []core.EnvVar{
					{
						Name:  "_CONF",
						Value: strings.Join(c, `\n`),
					},
				},
				ImagePullPolicy: core.PullIfNotPresent,
			},
		},
		RestartPolicy: core.RestartPolicyAlways,
	}
	pod.Labels = map[string]string{
		"_gateway": dns.Name,
	}
	_, pr, err := dnsc.PodClient.Post(pod)
	if err != nil {
		return err
	}
	svc := &core.Service{
		TypeMeta: meta.CreateTypeMeta(types.ServiceObjectType),
		ObjectMeta: meta.ObjectMeta{
			Name: "gateway-" + dns.Name,
		},
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				{
					Name:       "nginx",
					Port:       80,
					TargetPort: 80,
				},
			},
			Selector: map[string]string{
				"_gateway": dns.Name,
			},
			ClusterIP: dns.Spec.ServiceAddress,
			Type:      core.ServiceTypeClusterIP,
		},
		Status: core.ServiceStatus{},
	}
	_, sr, err := dnsc.ServiceClient.Post(svc)
	if err != nil {
		return err
	}

	s := 409
	for s != 200 {
		obj, err := dnsc.DnsClient.Get(dns.UID)
		if err != nil {
			return nil
		}
		dns = obj.(*core.DNS)
		dns.Status.PodUID = pr.UID
		dns.Status.ServiceUID = sr.UID
		//dns.SetResourceVersion(etcd.Rvm.GetNextResourceVersion())
		s, _, err = dnsc.DnsClient.Put(dns.UID, dns)
	}
	if err != nil {
		return err
	}
	//TODO delete service use uid saved in status
	//TODO delete pod use uid saved in status
	return nil
}

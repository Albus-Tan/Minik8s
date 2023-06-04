package pod

import (
	"context"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/types"
	client "minik8s/pkg/apiclient/interface"
	"minik8s/pkg/controller/cache"
	"minik8s/pkg/logger"
	"reflect"
	"time"
)

type PodController interface {
	Run(ctx context.Context)
}

func NewPodController(podClient client.Interface, podInformer cache.Informer) PodController {

	pc := &podController{
		Kind:        string(types.DnsObjectType),
		PodClient:   podClient,
		PodInformer: podInformer,
		restart:     cache.NewWorkQueue(),
	}

	_ = pc.PodInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: pc.updatePod,
	})

	return pc
}

type podController struct {
	Kind string

	PodInformer cache.Informer
	PodClient   client.Interface
	restart     cache.WorkQueue
}

func (pc *podController) Run(ctx context.Context) {

	go func() {
		logger.PodControllerLogger.Printf("[DnsController] start\n")
		defer logger.PodControllerLogger.Printf("[DnsController] finish\n")

		pc.runWorker(ctx)

		// wait for controller manager stop
		<-ctx.Done()
	}()
	return
}

func (pc *podController) PodKeyFunc(p *core.Pod) string {
	return p.GetUID()
}

func (pc *podController) enqueueRestart(p *core.Pod) {
	key := pc.PodKeyFunc(p)
	pc.restart.Enqueue(p)
	logger.PodControllerLogger.Printf("enqueueDNS uid %s\n", key)
}

func (pc *podController) updatePod(old, new interface{}) {
	o := old.(*core.Pod)
	n := new.(*core.Pod)
	logger.PodControllerLogger.Printf("Updating %s %s\n", pc.Kind, o.Name)
	if !reflect.DeepEqual(o.Spec, n.Spec) {
		return
	}
	if reflect.DeepEqual(o.Status, n.Status) {
		return
	}

	if o.Status.Phase != core.PodRunning {
		return
	}
	if o.Spec.RestartPolicy == core.RestartPolicyAlways {
		pc.enqueueRestart(o)
	}
	if o.Spec.RestartPolicy == core.RestartPolicyOnFailure && n.Status.Phase == core.PodFailed {
		pc.enqueueRestart(o)
	}

}

const defaultWorkerSleepInterval = time.Duration(3) * time.Second

func (pc *podController) runWorker(ctx context.Context) {
	// go wait.UntilWithContext(ctx, pc.worker, time.Second)
	for {
		select {
		case <-ctx.Done():
			logger.PodControllerLogger.Printf("[worker] ctx.Done() received, worker of PodController exit\n")
			return
		default:
			for pc.processNextWorkItem() {
			}
			time.Sleep(defaultWorkerSleepInterval)
		}
	}
}

func (pc *podController) processNextWorkItem() bool {

	item, ok := pc.restart.Dequeue()
	if !ok {
		return false
	}

	p := item.(*core.Pod)

	err := pc.processPodRestart(p)
	if err != nil {
		pc.restart.Enqueue(p)
		return false
	}

	return true
}

func (pc *podController) processPodRestart(p *core.Pod) error {
	_, _, err := pc.PodClient.Delete(p.UID)
	if err != nil {
		logger.PodControllerLogger.Printf("[processPodRestart] err: %v\n", err)
		return nil
	}
	_, _, err = pc.PodClient.Post(p)
	if err != nil {
		panic("Post failed")
	}
	return nil
}

package kubeproxy

import (
	"context"
	"errors"
	"log"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/types"
	"minik8s/pkg/api/watch"
	"minik8s/pkg/apiclient"
	client "minik8s/pkg/apiclient/interface"
	"minik8s/pkg/apiclient/listwatch"
	"minik8s/pkg/kubeproxy/service"
	"sync"
)

type KubeProxy interface {
	Run()
}

type kubeProxy struct {
	service.Manager

	podClient        client.Interface
	podListerWatcher listwatch.ListerWatcher
	svcClient        client.Interface
	svcListerWatcher listwatch.ListerWatcher

	sync.RWMutex
}

func New() KubeProxy {
	podClient, err := apiclient.NewRESTClient(types.PodObjectType)
	if err != nil {
		return nil
	}

	svcClient, err := apiclient.NewRESTClient(types.ServiceObjectType)
	if err != nil {
		return nil
	}

	return &kubeProxy{
		Manager:          service.New(),
		podClient:        podClient,
		podListerWatcher: listwatch.NewListWatchFromClient(podClient),
		svcClient:        svcClient,
		svcListerWatcher: listwatch.NewListWatchFromClient(svcClient),
		RWMutex:          sync.RWMutex{},
	}

}
func (k *kubeProxy) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	go k.watchPods(ctx, cancel)
	go k.watchSvcs(ctx, cancel)

	<-ctx.Done()
}

var (
	errorStopRequested = errors.New("stop requested")
)

func (k *kubeProxy) watchPods(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	w, err := k.podListerWatcher.Watch()
	if err != nil {
		log.Printf("[Kubelet] Watch pods error: %v\n", err)
	}

	err = k.handleWatchPods(w, ctx)
	w.Stop() // stop watch

	if err == errorStopRequested {
		return
	}

	if err != nil {
		log.Printf("[Kubelet] Watch pods error: %v\n", err)
	}
}

func (k *kubeProxy) watchSvcs(ctx context.Context, cancel context.CancelFunc) {
	defer cancel()
	w, err := k.svcListerWatcher.Watch()
	if err != nil {
		log.Printf("[Kubelet] Watch pods error: %v\n", err)
	}

	err = k.handleWatchSvcs(w, ctx)
	w.Stop() // stop watch

	if err == errorStopRequested {
		return
	}

	if err != nil {
		log.Printf("[Kubelet] Watch pods error: %v\n", err)
	}
}

func (k *kubeProxy) handleWatchPods(w watch.Interface, ctx context.Context) error {
	eventCount := 0
loop:
	for {
		select {
		case <-ctx.Done():
			log.Printf("[handleWatchPods]: ctx.Done(), Watch close - %v total %v items received\n", types.PodObjectType, eventCount)
			return errorStopRequested
		case event, ok := <-w.ResultChan():
			if !ok {
				break loop
			}

			p := event.Object.(*core.Pod)

			log.Printf("[handleWatchPods] event %v\n", event)
			log.Printf("[handleWatchPods] event object %v\n", event.Object)
			eventCount += 1

			switch event.Type {
			case watch.Added:
				// new Pod event, but not scheduled
				// ignore
			case watch.Modified:
				// Pod modified event
				// for multi machine, watch bind pod events belongs to watch.Modified event
				// check and distinguish create and update in handlePodModify func
				k.handlePodModify(p)
			case watch.Deleted:
				// Pod deleted event
				k.handlePodDel(p)

			case watch.Bookmark:
				panic("[handleWatchPods] Event Type watch.Bookmark received")
			case watch.Error:
				log.Printf("[handleWatchPods] watch.Error event object received %v\n", event.Object)
				log.Printf("[handleWatchPods]: Watch close - %v total %v items received\n", types.PodObjectType, eventCount)
				return event.Object.(*core.ErrorApiObject).GetError()
			default:
				panic("[handleWatchPods] Unknown Event Type received")
			}

		}
	}
	log.Printf("[handleWatchPods]: Watch close - %v total %v items received\n", types.PodObjectType, eventCount)
	return nil
}

func (k *kubeProxy) handleWatchSvcs(w watch.Interface, ctx context.Context) error {
	eventCount := 0
loop:
	for {
		select {
		case <-ctx.Done():
			log.Printf("[handleWatchSvcs]: ctx.Done(), Watch close - %v total %v items received\n", types.PodObjectType, eventCount)
			return errorStopRequested
		case event, ok := <-w.ResultChan():
			if !ok {
				break loop
			}

			s := event.Object.(*core.Service)

			log.Printf("[handleWatchSvcs] event %v\n", event)
			log.Printf("[handleWatchSvcs] event object %v\n", event.Object)
			eventCount += 1

			switch event.Type {
			case watch.Added:
				k.handleSvcCreat(s)
			case watch.Modified:
				// ignored
			case watch.Deleted:
				k.handleSvcDel(s)
			case watch.Bookmark:
				panic("[handleWatchSvcs] Event Type watch.Bookmark received")
			case watch.Error:
				log.Printf("[handleWatchSvcs] watch.Error event object received %v\n", event.Object)
				log.Printf("[handleWatchSvcs]: Watch close - %v total %v items received\n", types.PodObjectType, eventCount)
				return event.Object.(*core.ErrorApiObject).GetError()
			default:
				panic("[handleWatchSvcs] Unknown Event Type received")
			}

		}
	}
	log.Printf("[handleWatchPods]: Watch close - %v total %v items received\n", types.PodObjectType, eventCount)
	return nil
}

func (k *kubeProxy) handlePodModify(pod *core.Pod) {
	k.Lock()
	defer k.Unlock()
	k.Manager.HandlePodModify(pod)
}

func (k *kubeProxy) handlePodDel(pod *core.Pod) {
	k.Lock()
	defer k.Unlock()
	k.Manager.HandlePodDel(pod)
}

func (k *kubeProxy) handleSvcCreat(s *core.Service) {
	k.Lock()
	defer k.Unlock()
	k.Manager.CreatSvc(s)
}

func (k *kubeProxy) handleSvcDel(s *core.Service) {
	k.Lock()
	defer k.Unlock()
	k.Manager.DelSvc(s)
}

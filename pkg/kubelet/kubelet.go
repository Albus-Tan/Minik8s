package kubelet

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
	"minik8s/pkg/kubelet/pod"
)

type Kubelet interface {
	Run()
	Close()
}

func New() (Kubelet, error) {

	podClient, err := apiclient.NewRESTClient(types.PodObjectType)
	if err != nil {
		return nil, err
	}

	return &kubelet{
		name:             "Kubelet", // FIXME: change to node name + Kubelet
		podClient:        podClient,
		podListerWatcher: listwatch.NewListWatchFromClient(podClient),
		podManager:       pod.NewPodManager(),
	}, nil
}

func (k *kubelet) Close() {
}

type kubelet struct {
	name             string
	podClient        client.Interface
	podListerWatcher listwatch.ListerWatcher
	podManager       pod.Manager
}

func (k *kubelet) Run() {

	log.SetPrefix("[Kubelet] ")

	// use context to stop related go routines
	// after kubelet stop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start watch pods
	// FIXME: for multi machine, change it to watching bind pod events
	k.watchPods(ctx)
}

/*---------------------------- Watch Pods ----------------------------*/
var (
	errorStopRequested = errors.New("stop requested")
)

func (k *kubelet) watchPods(ctx context.Context) {

	log.Printf("[Kubelet] Start watch pods\n")

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

func (k *kubelet) handleWatchPods(w watch.Interface, ctx context.Context) error {
	eventCount := 0
loop:
	for {
		select {
		case <-ctx.Done():
			log.Printf("[handleWatchPods] %s: ctx.Done(), Watch close - %v total %v items received\n", k.name, types.PodObjectType, eventCount)
			return errorStopRequested
		case event, ok := <-w.ResultChan():
			if !ok {
				break loop
			}
			log.Printf("[handleWatchPods] event %v\n", event)
			log.Printf("[handleWatchPods] event object %v\n", event.Object)
			eventCount += 1

			p := event.Object.(*core.Pod)

			switch event.Type {
			case watch.Added:
				// new Pod event
				k.handlePodCreate(p)
			case watch.Modified:
				// Pod modified event
				k.handlePodModify(p)
			case watch.Deleted:
				// Pod deleted event
				k.handlePodDelete(p)
			case watch.Bookmark:
				panic("[handleWatchPods] Event Type watch.Bookmark received")
			case watch.Error:
				log.Printf("[handleWatchPods] watch.Error event object received %v\n", event.Object)
				log.Printf("[handleWatchPods] %s: Watch close - %v total %v items received\n", k.name, types.PodObjectType, eventCount)
				return event.Object.(*core.ErrorApiObject).GetError()
			default:
				panic("[handleWatchPods] Unknown Event Type received")
			}
		}
	}
	log.Printf("[handleWatchPods] %s: Watch close - %v total %v items received\n", k.name, types.PodObjectType, eventCount)
	return nil
}

func (k *kubelet) handlePodCreate(pod *core.Pod) {
	// FIXME: update container status field in pod

	// FIXME: update pod status

	// FIXME: send new pod status back to apiserver
	// add pod to podManager
	k.podManager.AddPod(pod)
}

func (k *kubelet) handlePodModify(pod *core.Pod) {
}

func (k *kubelet) handlePodDelete(pod *core.Pod) {
	// FIXME: update field in pod

	// FIXME: send new pod status back to apiserver

	// delete pod in podManager
	k.podManager.DeletePod(pod)
}

/*----------------------------  ----------------------------*/

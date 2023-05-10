package kubelet

import (
	"context"
	"errors"
	"log"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/watch"
	"minik8s/pkg/apiclient"
	client "minik8s/pkg/apiclient/interface"
	"minik8s/pkg/apiclient/listwatch"
	"minik8s/pkg/kubelet/constants"
	"minik8s/pkg/kubelet/container"
	"minik8s/pkg/kubelet/pod"
)

type Kubelet interface {
	Run()
}

func New() Kubelet {

	podClient, err := apiclient.NewRESTClient(core.PodObjectType)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = container.NewCriClient().CreateContainer(core.Container{Name: "hello", Image: "docker.io/library/redis:alpine"})
	if err != nil {
		log.Fatal(err.Error())
	}

	return &kubelet{
		name:             "Kubelet", // TODO: change to node name + Kubelet
		podClient:        podClient,
		podListerWatcher: listwatch.NewListWatchFromClient(podClient),
		podManager:       pod.NewPodManager(),
		criClient:        container.NewCriClient(),
	}
}

type kubelet struct {
	name             string
	podClient        client.Interface
	podListerWatcher listwatch.ListerWatcher
	podManager       pod.Manager
	criClient        container.CriClient
}

func (k *kubelet) Run() {

	log.SetPrefix("[Kubelet] ")

	// use context to stop related go routines
	// after kubelet stop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start watch pods
	// TODO: for multi machine, change it to watching bind pod events
	k.watchPods(ctx)

	//TODO
	panic("implement me")
}

/*---------------------------- Watch Pods ----------------------------*/
var (
	errorStopRequested = errors.New("stop requested")
)

func (k *kubelet) watchPods(ctx context.Context) {

	log.Printf("[Kubelet] Start watch pods\n")

	go func() {
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
	}()

}

func (k *kubelet) handleWatchPods(w watch.Interface, ctx context.Context) error {
	eventCount := 0
loop:
	for {
		select {
		case <-ctx.Done():
			log.Printf("[handleWatchPods] %s: ctx.Done(), Watch close - %v total %v items received\n", k.name, core.PodObjectType, eventCount)
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
				log.Printf("[handleWatchPods] %s: Watch close - %v total %v items received\n", k.name, core.PodObjectType, eventCount)
				return event.Object.(*core.ErrorApiObject).GetError()
			default:
				panic("[handleWatchPods] Unknown Event Type received")
			}
		}
	}
	log.Printf("[handleWatchPods] %s: Watch close - %v total %v items received\n", k.name, core.PodObjectType, eventCount)
	return nil
}

func (k *kubelet) handlePodCreate(pod *core.Pod) {
	// TODO: handle create new pod event

	// install Initial Containers
	pod.Spec.InitContainers = append(pod.Spec.InitContainers, constants.InitialPauseContainer)

	for _, c := range pod.Spec.InitContainers {
		err := k.criClient.CreateContainer(c)
		if err != nil {
			log.Printf("[handlePodCreate] create init container %v failed: %v", c.Name, err)
			return
		}
	}

	// TODO: run container and update field in pod
	for _, c := range pod.Spec.Containers {
		err := k.criClient.CreateContainer(c)
		if err != nil {
			log.Printf("[handlePodCreate] create container %v failed: %v", c.Name, err)
			return
		}
	}

	// TODO: update pod status

	// TODO: send new pod status back to apiserver

	// add pod to podManager
	k.podManager.AddPod(pod)
}

func (k *kubelet) handlePodModify(pod *core.Pod) {
	// TODO: handle pod modified in apiserver etcd

	// TODO: process actual update and update field in pod

	// TODO: update pod status

	// TODO: send new pod status back to apiserver

	// update pod in podManager
	k.podManager.UpdatePod(pod)
}

func (k *kubelet) handlePodDelete(pod *core.Pod) {
	// TODO: handle pod deleted in apiserver etcd

	// TODO: process actual delete such as container delete and update field in pod
	for idx, c := range pod.Spec.Containers {
		err := k.criClient.CleanContainer(c.Name) // TODO @wjr: I'm not sure whether I give the correct param here
		if err != nil {
			log.Printf("[handlePodCreate] clean container %v failed: %v", pod.Spec.Containers[idx].Name, err)
			return
		}
	}

	// TODO: update pod status

	// TODO: send new pod status back to apiserver

	// delete pod in podManager
	k.podManager.DeletePod(pod)
}

/*----------------------------  ----------------------------*/

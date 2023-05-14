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
	"minik8s/pkg/kubelet/constants"
	"minik8s/pkg/kubelet/container/cri"
	"minik8s/pkg/kubelet/pod"
	"sync"
	"time"
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

	criClient, err := cri.NewDocker()
	if err != nil {
		return nil, err
	}

	return &kubelet{
		name:             "Kubelet", // FIXME: change to node name + Kubelet
		podClient:        podClient,
		podListerWatcher: listwatch.NewListWatchFromClient(podClient),
		podManager:       pod.NewPodManager(),
		criClient:        criClient,
	}, nil
}

func (k *kubelet) Close() {
}

type kubelet struct {
	name             string
	podClient        client.Interface
	podListerWatcher listwatch.ListerWatcher
	podManager       pod.Manager
	criClient        cri.Client
	lock             sync.RWMutex
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
	k.lock.Lock()
	defer k.lock.Unlock()
	ctx, cancel := context.WithCancel(context.Background())
	k.createMasterContainer(ctx, pod)
	k.createContainers(ctx, pod, pod.Spec.Containers)
	go k.startWatchContainers(ctx, *pod)
	pod.CancelWorker = cancel
	// add pod to podManager
	k.podManager.AddPod(pod)
}

func (k *kubelet) handlePodModify(pod *core.Pod) {
	k.lock.Lock()
	defer k.lock.Unlock()
	old, found := k.podManager.GetPodByUID(pod.UID)
	if !found {
		k.handlePodCreate(pod)
		return
	}
	up := containersNew(old.Spec.Containers, pod.Spec.Containers)
	down := containersNew(pod.Spec.Containers, old.Spec.Containers)
	old.CancelWorker()

	ctx, cancel := context.WithCancel(context.Background())
	pod.CancelWorker = cancel
	k.removeContainers(ctx, pod, down)
	k.createContainers(ctx, pod, up)
	k.podManager.UpdatePod(pod)
}

func (k *kubelet) handlePodDelete(pod *core.Pod) {
	k.lock.Lock()
	defer k.lock.Unlock()
	old, find := k.podManager.GetPodByUID(pod.UID)
	if !find {
		log.Println("unconsistent delete")
		return
	}
	old.CancelWorker()
	ctx := context.Background()
	k.removeContainers(ctx, old, old.Spec.Containers)
	k.removeMasterContainer(ctx, pod)

	// delete pod in podManager
	k.podManager.DeletePod(old)
}

/*----------------------------  ----------------------------*/

func containersNew(old []core.Container, new []core.Container) []core.Container {
	set := make(map[string]core.Container)
	for _, c := range new {
		set[c.Name] = c
	}

	for _, c := range old {
		delete(set, c.Name)
	}

	ret := make([]core.Container, 0)
	for _, c := range set {
		ret = append(ret, c)
	}

	return ret
}

func createPodContainerName(pod *core.Pod, container core.Container) string {
	return pod.UID + "-" + container.Name
}

func (k *kubelet) createMasterContainer(ctx context.Context, pod *core.Pod) {
	container := constants.InitialPauseContainer
	container.Name = createPodContainerName(pod, container)
	_, err := k.criClient.ContainerCreate(ctx, container)
	if err != nil {
		log.Fatalf("create failed")
	}

	if err := k.criClient.ContainerStart(ctx, container.Name); err != nil {
		log.Fatalf("run failed")
	}
}

func (k *kubelet) createContainers(ctx context.Context, pod *core.Pod, containers []core.Container) {
	for _, container := range containers {
		name := container.Name
		container.Name = createPodContainerName(pod, container)
		container.Master = k.criClient.ContainerId(ctx, createPodContainerName(pod, constants.InitialPauseContainer))
		if container.Master == "" {
			log.Fatalf("MissingMaster")
		}
		id, err := k.criClient.ContainerCreate(ctx, container)
		if err != nil {
			log.Fatalf("create failed")
		}
		pod.Status.ContainerStatuses = append(pod.Status.ContainerStatuses, core.ContainerStatus{
			Name: name,
			State: core.ContainerState{
				Waiting:    nil,
				Running:    &core.ContainerStateRunning{},
				Terminated: nil,
			},
			Image:       container.Image,
			ImageID:     container.Image,
			ContainerID: id,
		})
		if err := k.criClient.ContainerStart(ctx, container.Name); err != nil {
			log.Fatalf("run failed")
		}
	}
}

func (k *kubelet) removeContainers(ctx context.Context, pod *core.Pod, containers []core.Container) {
	for _, container := range containers {
		container.Name = createPodContainerName(pod, container)
		container.Master = createPodContainerName(pod, constants.InitialPauseContainer)
		err := k.criClient.ContainerRemove(ctx, container.Name)
		if err != nil {
			log.Println("[ERROR]: failed to remove container", container.Name, err.Error())
		}
	}
}

func (k *kubelet) removeMasterContainer(ctx context.Context, pod *core.Pod) {
	err := k.criClient.ContainerRemove(
		ctx,
		createPodContainerName(pod, constants.InitialPauseContainer),
	)

	if err != nil {
		log.Println("[ERROR]: failed to remove pod master container ", pod.Name, err.Error())
	}
}

func (k *kubelet) startWatchContainers(ctx context.Context, pod core.Pod) {
	pod.Status.Phase = core.PodRunning
	log.Println("start watch {}", pod.UID)
	for {
		select {
		case <-ctx.Done():
			log.Println("stop watch {} ", pod.UID)
			return
		default:
			k.lock.Lock()
			changed := false
			for idx, c := range pod.Status.ContainerStatuses {
				if c.State.Running != nil {
					r, err := k.criClient.ContainerInspect(ctx, c.ContainerID)
					if err != nil {
						log.Println(err.Error())
					}
					if !r || err != nil {
						changed = true
						pod.Status.ContainerStatuses[idx].State = core.ContainerState{Terminated: &core.ContainerStateTerminated{
							ExitCode:    0,
							Signal:      0,
							Reason:      "",
							Message:     "",
							ContainerID: c.ContainerID,
						}}
					}
				}
			}
			if changed {
				_, _, err := k.podClient.Put(pod.UID, &pod)
				if err != nil {
					log.Println(err.Error())

				}
				k.lock.Unlock()
				return
			}
			k.lock.Unlock()
			time.Sleep(time.Second)
		}
	}

}

package kubelet

import (
	"context"
	"errors"
	"log"
	"minik8s/config"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/types"
	"minik8s/pkg/api/watch"
	"minik8s/pkg/apiclient"
	client "minik8s/pkg/apiclient/interface"
	"minik8s/pkg/apiclient/listwatch"
	"minik8s/pkg/cadvisor"
	"minik8s/pkg/kubelet/constants"
	"minik8s/pkg/kubelet/container/cri"
	"minik8s/pkg/kubelet/pod"
	"minik8s/pkg/logger"
	"reflect"
	"sync"
	"time"
)

type Kubelet interface {
	Run()
	Close()
}

func New(node *core.Node) (Kubelet, error) {

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
		cadvisorClient:   cadvisor.NewClient(config.CadvisorUrl(config.CadvisorHost)),
		node:             node,
	}, nil
}

func (k *kubelet) Close() {
}

type kubelet struct {
	name             string
	node             *core.Node
	podClient        client.Interface
	podListerWatcher listwatch.ListerWatcher
	podManager       pod.Manager
	criClient        cri.Client
	lock             sync.RWMutex
	cadvisorClient   cadvisor.Interface
}

func (k *kubelet) Run() {

	log.SetPrefix("[Kubelet] ")

	// use context to stop related go routines
	// after kubelet stop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start cadvisor on current node
	k.startCadvisorClient()

	k.listPods(ctx)

	// start watch pods
	k.watchPods(ctx)
}

/*---------------------------- cadvisor ----------------------------*/
func (k *kubelet) startCadvisorClient() {
	log.Printf("[Kubelet] Start cadvisor client\n")
	err := k.cadvisorClient.Start()
	if err != nil {
		log.Printf("[Kubelet] Start cadvisor error: %v\n", err)
		return
	}
}

/*---------------------------- Watch Pods ----------------------------*/
var (
	errorStopRequested = errors.New("stop requested")
)

func (k *kubelet) listPods(ctx context.Context) {

	log.Printf("[Kubelet] Start list pods\n")

	podList, err := k.podListerWatcher.List()
	if err != nil {
		log.Fatalln(err)
	}

	for _, p := range podList.GetIApiObjectArr() {
		k.podManager.UpdatePod(p.(*core.Pod))
	}
	for _, p := range podList.GetIApiObjectArr() {
		go k.startWatchContainers(ctx, *p.(*core.Pod))
	}
}

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

			p := event.Object.(*core.Pod)

			// filter pod not belong to current node
			if p.Spec.NodeName == k.node.Name {

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
	}
	log.Printf("[handleWatchPods] %s: Watch close - %v total %v items received\n", k.name, types.PodObjectType, eventCount)
	return nil
}

func (k *kubelet) handlePodModify(pod *core.Pod) {
	k.lock.Lock()
	defer k.lock.Unlock()
	old, found := k.podManager.GetPodByUID(pod.UID)
	if !found {
		k.createPod(pod)
		return
	}

	if reflect.DeepEqual(old.Spec, pod.Spec) {
		return
	}

	logger.KubeletLogger.Printf("Pod %v update on current node %v, start handle pod modify\n", pod.UID, k.node.Name)

	up := containersNew(old.Spec.Containers, pod.Spec.Containers)
	down := containersNew(pod.Spec.Containers, old.Spec.Containers)

	ctx := context.Background()
	k.removeContainers(ctx, pod, down)
	k.createContainers(ctx, pod, up)
	go k.startWatchContainers(ctx, *pod)
}

func (k *kubelet) handlePodDelete(pod *core.Pod) {
	k.lock.Lock()
	defer k.lock.Unlock()

	logger.KubeletLogger.Printf("Pod %v delete on current node %v, start handle pod delete\n", pod.UID, k.node.Name)

	old, find := k.podManager.GetPodByUID(pod.UID)
	if !find {
		log.Println("unconsistent delete")
		return
	}
	ctx := context.Background()
	k.removeContainers(ctx, old, old.Spec.Containers)
	k.removeMasterContainer(ctx, pod)

	// delete pod in podManager
	k.podManager.DeletePod(old)
}

/*----------------------------  ----------------------------*/

func (k *kubelet) createPod(pod *core.Pod) {

	logger.KubeletLogger.Printf("New Pod %v bind to current node %v, start handle pod create on current machine\n", pod.UID, k.node.Name)

	ctx := context.Background()
	k.createMasterContainer(ctx, pod)
	k.createContainers(ctx, pod, pod.Spec.Containers)
	go k.startWatchContainers(ctx, *pod)
	// add pod to podManager
	k.podManager.AddPod(pod)
}

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

func makePodContainerName(pod *core.Pod, container core.Container) string {
	return pod.UID + "-" + container.Name
}

func (k *kubelet) createMasterContainer(ctx context.Context, pod *core.Pod) {
	container := constants.InitialPauseContainer
	container.Name = makePodContainerName(pod, container)
	_, err := k.criClient.ContainerCreate(ctx, container)
	if err != nil {
		log.Fatalf("create failed %v", err)
	}

	if err := k.criClient.ContainerStart(ctx, container.Name); err != nil {
		log.Fatalf("run failed")
	}
}

func (k *kubelet) createContainers(ctx context.Context, pod *core.Pod, containers []core.Container) {
	for _, container := range containers {
		name := container.Name
		container.Name = makePodContainerName(pod, container)
		container.Master = k.criClient.ContainerId(ctx, makePodContainerName(pod, constants.InitialPauseContainer))
		if container.Master == "" {
			log.Fatalf("MissingMaster")
		}
		id, err := k.criClient.ContainerCreate(ctx, container)
		if err != nil {
			log.Fatalf("create failed %v", err)
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
		container.Name = makePodContainerName(pod, container)
		container.Master = makePodContainerName(pod, constants.InitialPauseContainer)
		err := k.criClient.ContainerRemove(ctx, container.Name)
		if err != nil {
			log.Println("[ERROR]: failed to remove container", container.Name, err.Error())
		}
	}
}

func (k *kubelet) removeMasterContainer(ctx context.Context, pod *core.Pod) {
	err := k.criClient.ContainerRemove(
		ctx,
		makePodContainerName(pod, constants.InitialPauseContainer),
	)

	if err != nil {
		log.Println("[ERROR]: failed to remove pod master container ", pod.Name, err.Error())
	}
}

func (k *kubelet) startWatchContainers(ctx context.Context, pod core.Pod) {
	for {
		err := k.inspectContainer(ctx, pod)
		if err != nil {
			logger.KubeletLogger.Printf("%s stale", pod.Name)
			return
		}
		time.Sleep(time.Second)
	}
}

func (k *kubelet) inspectContainer(ctx context.Context, pod core.Pod) error {
	ip, err := k.criClient.ContainerIP(ctx, k.criClient.ContainerId(ctx, pod.UID+"-"+"pause"))

	if err != nil {
		return err
	}
	ncs := make([]core.ContainerStatus, 0)
	for _, container := range pod.Spec.Containers {
		ncs = append(ncs, core.ContainerStatus{
			Name: container.Name,
			State: core.ContainerState{
				Waiting:    nil,
				Running:    &core.ContainerStateRunning{},
				Terminated: nil,
			},
			Image:       container.Image,
			ImageID:     container.Image,
			ContainerID: k.criClient.ContainerId(ctx, pod.UID+"-"+container.Name),
		})
	}
	for idx, c := range ncs {
		r, e, err := k.criClient.ContainerStatus(ctx, c.ContainerID)
		if err != nil {
			return err
		}
		if !r || err != nil {
			ncs[idx].State = core.ContainerState{
				Terminated: &core.ContainerStateTerminated{
					ExitCode:    int32(e),
					Signal:      0,
					Reason:      "",
					Message:     "",
					ContainerID: c.ContainerID,
				},
			}
		}
	}
	var ns core.PodPhase
	t := true
	for _, c := range ncs {
		if c.State.Running != nil {
			t = false
		}
	}
	if t {
		s := true
		for _, c := range ncs {
			if c.State.Terminated.ExitCode != 0 {
				s = false
			}
		}
		if s {
			ns = core.PodSucceeded
		} else {
			ns = core.PodFailed
		}
	} else {
		ns = core.PodRunning
	}

	if ns != pod.Status.Phase || ip != pod.Status.PodIP || !reflect.DeepEqual(ncs, pod.Status.ContainerStatuses) {
		pod.Status.Phase = ns
		pod.Status.PodIP = ip
		pod.Status.ContainerStatuses = ncs
		r, err := k.podClient.Get(pod.UID)
		if err != nil {
			return err
		}
		rr := r.(*core.Pod)
		if reflect.DeepEqual(rr.Spec, pod.Spec) && !reflect.DeepEqual(rr.Status, pod.Status) {
			rr.Status = pod.Status
			_, _, err = k.podClient.Put(pod.UID, rr)
		}
	}
	return nil
}

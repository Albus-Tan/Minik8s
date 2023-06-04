package scheduler

import (
	"errors"
	"golang.org/x/net/context"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/types"
	"minik8s/pkg/api/watch"
	"minik8s/pkg/apiclient"
	client "minik8s/pkg/apiclient/interface"
	"minik8s/pkg/apiclient/listwatch"
	"minik8s/pkg/logger"
	"minik8s/pkg/node"
	"minik8s/utils/datastructure"
	"net/http"
	"time"
)

// Scheduler watches for new unscheduled pods. It attempts to find
// nodes that they fit on and writes bindings back to the api server.
type Scheduler struct {

	// Client
	podClient  client.Interface
	nodeClient client.Interface

	// ListWatcher
	podListWatcher  listwatch.ListerWatcher
	nodeListWatcher listwatch.ListerWatcher

	// Close this to shut down the scheduler.
	StopEverything <-chan struct{}

	// schedulingQueue holds pods to be scheduled
	schedulingQueue datastructure.IConcurrentQueue

	// nodesQueue holds nodes to be scheduled to for RR
	nodesQueue datastructure.IConcurrentQueue
}

func NewScheduler() *Scheduler {

	podClient, _ := apiclient.NewRESTClient(types.PodObjectType)
	podListWatcher := listwatch.NewListWatchFromClient(podClient)
	nodeClient, _ := apiclient.NewRESTClient(types.NodeObjectType)
	nodeListWatcher := listwatch.NewListWatchFromClient(nodeClient)

	return &Scheduler{
		podClient:       podClient,
		podListWatcher:  podListWatcher,
		nodeClient:      nodeClient,
		nodeListWatcher: nodeListWatcher,
		schedulingQueue: datastructure.NewConcurrentQueue(),
		nodesQueue:      datastructure.NewConcurrentQueue(),
	}
}

func (s *Scheduler) Run(ctx context.Context, cancel context.CancelFunc) {

	logger.SchedulerLogger.Printf("[Scheduler] start\n")
	defer logger.SchedulerLogger.Printf("[Scheduler] init finish\n")

	syncChan := make(chan bool)

	go func() {
		defer cancel()
		err := s.listAndWatchNodes(syncChan, ctx.Done())
		if err != nil {
			logger.SchedulerLogger.Printf("[Scheduler] listAndWatchNodes failed, err: %v\n", err)
		}
	}()

	// wait for node list finish
	<-syncChan

	go func() {
		defer cancel()
		err := s.listAndWatchPods(ctx.Done())
		if err != nil {
			logger.SchedulerLogger.Printf("[Scheduler] listAndWatchPods failed, err: %v\n", err)
		}
	}()

	go func() {
		defer cancel()
		s.runScheduleWorker(ctx)
	}()

}

const defaultWorkerSleepInterval = time.Duration(3) * time.Second

func (s *Scheduler) runScheduleWorker(ctx context.Context) {

	// go wait.UntilWithContext(ctx, rsc.worker, time.Second)
	for {
		select {
		case <-ctx.Done():
			logger.SchedulerLogger.Printf("[worker] ctx.Done() received, worker of Scheduler exit\n")
			return
		default:
			for s.processNextPodToSchedule() {
			}
			time.Sleep(defaultWorkerSleepInterval)
		}
	}

}

func (s *Scheduler) processNextPodToSchedule() bool {

	pod := s.dequeuePod()
	if pod == nil {
		return false
	}

	nodeBind := s.doSchedule(pod)
	if nodeBind == nil {
		logger.SchedulerLogger.Printf("[processNextPodToSchedule] pod %v bind failed\n", pod.UID)
		s.enqueuePod(pod)
		return false
	}

	// modify pod.Spec.NodeName
	pod.Spec.NodeName = nodeBind.Name

	// send binding result to apiserver
	code, _, err := s.podClient.Put(pod.UID, pod)
	if err != nil {
		for code == http.StatusConflict {
			podItem, _ := s.podClient.Get(pod.UID)
			pod = podItem.(*core.Pod)
			pod.Spec.NodeName = nodeBind.Name
			code, _, err = s.podClient.Put(pod.UID, pod)
		}
		return code == http.StatusOK
	}

	logger.SchedulerLogger.Printf("[processNextPodToSchedule] schedule pod uid %v to node %v\n", pod.UID, nodeBind.Name)

	return true
}

var (
	errorStopRequested = errors.New("stop requested")
)

func (s *Scheduler) listAndWatchNodes(syncChan chan bool, stopCh <-chan struct{}) error {

	// list all nodes and push into nodesQueue
	nodesList, err := s.nodeListWatcher.List()
	if err != nil {
		return err
	}

	nodeItems := nodesList.GetIApiObjectArr()
	for _, item := range nodeItems {
		no := item.(*core.Node)
		s.nodesQueue.Enqueue(no)
	}

	// send signal through syncChan to tell scheduler list node finish
	syncChan <- true

	// start watch nodes change
	var w watch.Interface
	w, err = s.nodeListWatcher.Watch()
	if err != nil {
		return err
	}

	err = s.handleWatchNodes(w, stopCh)
	w.Stop() // stop watch

	if err == errorStopRequested {
		return nil
	}

	return err

}

func (s *Scheduler) listAndWatchPods(stopCh <-chan struct{}) error {

	// list all pods and push into podsQueue
	podsList, err := s.podListWatcher.List()
	if err != nil {
		return err
	}

	podItems := podsList.GetIApiObjectArr()
	for _, item := range podItems {
		pod := item.(*core.Pod)
		s.enqueuePod(pod)
	}

	// start watch pods change
	var w watch.Interface
	w, err = s.podListWatcher.Watch()
	if err != nil {
		return err
	}

	err = s.handleWatchPods(w, stopCh)
	w.Stop() // stop watch

	if err == errorStopRequested {
		return nil
	}

	return err

}

func (s *Scheduler) enqueuePod(pod *core.Pod) {
	s.schedulingQueue.Enqueue(pod)
	logger.SchedulerLogger.Printf("[enqueuePod] pod %v enqueued\n", pod.UID)
}

func (s *Scheduler) dequeuePod() *core.Pod {
	podItem, exist := s.schedulingQueue.Dequeue()
	if exist {
		pod := podItem.(*core.Pod)
		logger.SchedulerLogger.Printf("[dequeuePod] pod %v equeued\n", pod.UID)
		return pod
	} else {
		logger.SchedulerLogger.Printf("[dequeuePod] queue empty\n")
		return nil
	}
}

func (s *Scheduler) handleWatchPods(w watch.Interface, stopCh <-chan struct{}) error {
	eventCount := 0
loop:
	for {
		select {
		case <-stopCh:
			return errorStopRequested
		case event, ok := <-w.ResultChan():
			if !ok {
				break loop
			}
			logger.SchedulerLogger.Printf("[handleWatchPods] event %v\n", event)
			logger.SchedulerLogger.Printf("[handleWatchPods] event object %v\n", event.Object)
			eventCount += 1

			switch event.Type {
			case watch.Added:
				newPod := (event.Object).(*core.Pod)
				s.enqueuePod(newPod)
				logger.SchedulerLogger.Printf("[handleWatchPods] new Pod event, handle pod %v created\n", newPod.UID)
			case watch.Modified:
				// ignore
			case watch.Deleted:
				// ignore
			case watch.Bookmark:
				panic("[handleWatchPods] watchHandler Event Type watch.Bookmark received")
			case watch.Error:
				panic("[handleWatchPods] watchHandler Event Type watch.Error received")
			default:
				panic("[handleWatchPods] watchHandler Unknown Event Type received")
			}
		}
	}
	return nil
}

func (s *Scheduler) handleWatchNodes(w watch.Interface, stopCh <-chan struct{}) error {
	eventCount := 0
loop:
	for {
		select {
		case <-stopCh:
			return errorStopRequested
		case event, ok := <-w.ResultChan():
			if !ok {
				break loop
			}
			logger.SchedulerLogger.Printf("[handleWatchNodes] event %v\n", event)
			logger.SchedulerLogger.Printf("[handleWatchNodes] event object %v\n", event.Object)
			eventCount += 1

			switch event.Type {
			case watch.Added:
				newNode := (event.Object).(*core.Node)
				if newNode.Status.Phase == core.NodeRunning {
					s.nodesQueue.Enqueue(newNode)
				}
			case watch.Modified:
				newNode := (event.Object).(*core.Node)
				if newNode.Status.Phase == core.NodeRunning {
					nodeUID := newNode.GetUID()
					s.deleteNodeInQueue(nodeUID)
					s.nodesQueue.Enqueue(newNode)
				}
			case watch.Deleted:
				oldNode := (event.Object).(*core.Node)
				nodeUID := oldNode.GetUID()
				s.deleteNodeInQueue(nodeUID)
			case watch.Bookmark:
				panic("[handleWatchNodes] watchHandler Event Type watch.Bookmark received")
			case watch.Error:
				panic("[handleWatchNodes] watchHandler Event Type watch.Error received")
			default:
				panic("[handleWatchNodes] watchHandler Unknown Event Type received")
			}
		}
	}
	return nil
}

func (s *Scheduler) deleteNodeInQueue(uid types.UID) bool {
	allNodes := s.nodesQueue.GetContent()
	for i, n := range allNodes {
		no := n.(*core.Node)
		if uid == no.UID {
			newNodes := append(allNodes[:i], allNodes[i+1:]...)
			s.nodesQueue.SetContent(newNodes)
			return true
		}
	}
	return false
}

func (s *Scheduler) getNodeInQueue(name string) *core.Node {
	allNodes := s.nodesQueue.GetContent()
	for _, n := range allNodes {
		no := n.(*core.Node)
		if name == no.Name {
			return no
		}
	}
	return nil
}

// doSchedule do not schedule pod to master
func (s *Scheduler) doSchedule(newPod *core.Pod) *core.Node {

	if s.nodesQueue.Length() == 1 {
		n, _ := s.nodesQueue.Front()
		if n.(*core.Node).Name == node.NameMaster {
			panic("[doSchedule] only have master node, can not schedule!\n")
		}
	}

	if newPod.Spec.NodeName != "" {
		return s.doScheduleNodeAffinity(newPod)
	} else {
		var nodeScheduled *core.Node = nil
		if newPod.Spec.Affinity != nil {
			nodeScheduled = s.doSchedulePodAntiAffinity(newPod)
		}
		if nodeScheduled == nil {
			nodeScheduled = s.doScheduleRR()
		}
		return nodeScheduled
	}
}

func (s *Scheduler) moveNodeToEnd(nodeName string) {
	no := s.getNodeInQueue(nodeName)
	if no != nil {
		if s.deleteNodeInQueue(no.UID) {
			s.nodesQueue.Enqueue(no)
		}
	}
}

func (s *Scheduler) doScheduleRR() *core.Node {
	if s.nodesQueue.Empty() {
		logger.SchedulerLogger.Printf("[Scheduler][doSchedule] no nodes registered in queue\n")
		return nil
	}
	item, _ := s.nodesQueue.Dequeue()
	n := item.(*core.Node)
	for n.Name == node.NameMaster {
		s.nodesQueue.Enqueue(item)
		item, _ = s.nodesQueue.Dequeue()
		n = item.(*core.Node)
	}
	s.nodesQueue.Enqueue(item)
	return n
}

// NodeName is a request to schedule this pod onto a specific node. If it is non-empty,
// the scheduler simply schedules this pod onto that node, assuming that it fits resource
// requirements.
func (s *Scheduler) doScheduleNodeAffinity(newPod *core.Pod) *core.Node {
	return s.getNodeInQueue(newPod.Spec.NodeName)
}

func (s *Scheduler) doSchedulePodAntiAffinity(newPod *core.Pod) *core.Node {

	logger.SchedulerLogger.Printf("[Scheduler][doSchedulePodAntiAffinity] start scheduling pod %v, uid %v\n", newPod.Name, newPod.UID)

	newPodAnti := newPod.Spec.Affinity.PodAntiAffinity
	var nodeNameNotToSchedule []string

	// Get all pods scheduled already
	podList, _ := s.podListWatcher.List()
	for _, podItem := range podList.GetIApiObjectArr() {
		pod := podItem.(*core.Pod)
		if pod.UID == newPod.UID {
			continue
		}
		// check affinity
		for _, podAffinityTerm := range newPodAnti.RequiredDuringSchedulingIgnoredDuringExecution {
			if podAffinityTerm.LabelSelector != nil {
				logger.SchedulerLogger.Printf("[Scheduler][doSchedulePodAntiAffinity] new pod label selector: %v, pod label %v\n", (*podAffinityTerm.LabelSelector).MatchLabels, pod.Labels)
				isMatched := meta.MatchLabelSelector(*podAffinityTerm.LabelSelector, pod.Labels)
				if isMatched {
					logger.SchedulerLogger.Printf("[Scheduler][doSchedulePodAntiAffinity] new pod can not schedule to node %v due to pod %v, uid %v\n", pod.Spec.NodeName, pod.Name, pod.UID)
					nodeNameNotToSchedule = append(nodeNameNotToSchedule, pod.Spec.NodeName)
				}
			}
		}
	}

	if len(nodeNameNotToSchedule) == 0 {
		logger.SchedulerLogger.Printf("[Scheduler][doSchedulePodAntiAffinity] no node can not be schedule, use rr\n")
		return nil
	} else {
		allNodes := s.nodesQueue.GetContent()
		for _, n := range allNodes {
			no := n.(*core.Node)
			notSchedule := false
			logger.SchedulerLogger.Printf("[Scheduler][doSchedulePodAntiAffinity] checking node %v\n", no.Name)
			for _, nodeNotSchedule := range nodeNameNotToSchedule {
				if nodeNotSchedule == no.Name {
					logger.SchedulerLogger.Printf("[Scheduler][doSchedulePodAntiAffinity] pod anti-affinity: no schedule to node %v\n", no.Name)
					notSchedule = true
					break
				}
			}
			if !notSchedule {
				// check if is master
				if no.Name == node.NameMaster {
					logger.SchedulerLogger.Printf("[Scheduler][doSchedulePodAntiAffinity] current node master\n", no.Name)
					continue
				}
				// do schedule
				logger.SchedulerLogger.Printf("[Scheduler][doSchedulePodAntiAffinity] pod anti-affinity schedule success to node %v\n", no.Name)
				// put node scheduled to last of rr queue
				s.moveNodeToEnd(no.Name)
				return no
			}
		}
	}
	logger.SchedulerLogger.Printf("[Scheduler][doSchedulePodAntiAffinity] no nodes left to satisfy pod anti-affinity in queue, use rr instead\n")
	return nil
}

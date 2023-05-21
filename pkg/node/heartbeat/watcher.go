package heartbeat

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
	"minik8s/pkg/logger"
	"minik8s/pkg/node"
	"sync"
	"time"
)

type Watcher interface {
	Run(ctx context.Context, cancel context.CancelFunc)
}

func NewWatcher() Watcher {
	nodeCli, _ := apiclient.NewRESTClient(types.NodeObjectType)
	nodeListWatcher := listwatch.NewListWatchFromClient(nodeCli)

	return &watcher{
		nodeClient:           nodeCli,
		nodeListWatcher:      nodeListWatcher,
		lastHeartbeatTimeMap: nil,
	}
}

type watcher struct {
	nodeClient      client.Interface
	nodeListWatcher listwatch.ListerWatcher

	lastHeartbeatTimeMap map[types.UID]time.Time // master should not be record in this map
	heartbeatTimeMapLock sync.RWMutex
}

func (w *watcher) Run(ctx context.Context, cancel context.CancelFunc) {
	log.Printf("[HeartbeatWatcher] start\n")
	defer log.Printf("[HeartbeatWatcher] running\n")

	syncChan := make(chan bool)

	go func() {
		defer cancel()
		defer log.Printf("[HeartbeatWatcher] finished\n")
		err := w.listAndWatchNodes(syncChan, ctx.Done())
		if err != nil {
			log.Printf("[HeartbeatWatcher] listAndWatchNodes failed, err: %v\n", err)
		}
	}()

	// wait for node list finish
	<-syncChan

	// run worker to delete node whose heartbeat have not been received for long
	go func() {
		defer cancel()
		w.checkHeartbeat(ctx)
	}()
}

const defaultWorkerSleepInterval = config.HeartbeatCheckInterval

func (w *watcher) checkHeartbeat(ctx context.Context) {

	// go wait.UntilWithContext(ctx, rsc.worker, time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Printf("[HeartbeatWatcher] ctx.Done() received, worker of Heartbeat Watcher exit\n")
			return
		default:

			// check heartbeat time
			w.heartbeatTimeMapLock.RLock()
			for uid, t := range w.lastHeartbeatTimeMap {
				if time.Since(t) > config.HeartbeatDeadInterval {
					log.Printf("[HeartbeatWatcher] checkHeartbeat: node %v dead, last heartbeat time %v, current time %v\n", uid, t, time.Now())
					// node dead, delete that node
					_, _, err := w.nodeClient.Delete(uid)
					if err != nil {
						log.Printf("[HeartbeatWatcher] checkHeartbeat: node %v delete failed\n", uid)
					}
				}
			}
			w.heartbeatTimeMapLock.RUnlock()

			time.Sleep(defaultWorkerSleepInterval)
		}
	}

}

var (
	errorStopRequested = errors.New("stop requested")
)

func (w *watcher) isMaster(no *core.Node) bool {
	return no.Name == node.NameMaster
}

func (w *watcher) listAndWatchNodes(syncChan chan bool, stopCh <-chan struct{}) error {

	// list all nodes
	nodesList, err := w.nodeListWatcher.List()
	if err != nil {
		return err
	}

	nodeItems := nodesList.GetIApiObjectArr()

	w.heartbeatTimeMapLock.Lock()

	// create map (without master)
	w.lastHeartbeatTimeMap = make(map[types.UID]time.Time, len(nodeItems)-1)

	for _, item := range nodeItems {
		no := item.(*core.Node)
		if !w.isMaster(no) {
			w.lastHeartbeatTimeMap[no.UID] = time.Now()
		}
	}

	w.heartbeatTimeMapLock.Unlock()

	// send signal through syncChan to tell list node finish
	syncChan <- true

	// start watch nodes change
	var wi watch.Interface
	wi, err = w.nodeListWatcher.Watch()
	if err != nil {
		return err
	}

	err = w.handleWatchNodes(wi, stopCh)
	wi.Stop() // stop watch

	if err == errorStopRequested {
		return nil
	}

	return err

}

func (w *watcher) handleWatchNodes(wi watch.Interface, stopCh <-chan struct{}) error {
	eventCount := 0
loop:
	for {
		select {
		case <-stopCh:
			return errorStopRequested
		case event, ok := <-wi.ResultChan():
			if !ok {
				break loop
			}
			logger.SchedulerLogger.Printf("[handleWatchNodes] event %v\n", event)
			logger.SchedulerLogger.Printf("[handleWatchNodes] event object %v\n", event.Object)
			eventCount += 1

			switch event.Type {
			case watch.Added:
				// add new node to map
				newNode := (event.Object).(*core.Node)
				if !w.isMaster(newNode) {
					w.heartbeatTimeMapLock.Lock()
					w.lastHeartbeatTimeMap[newNode.UID] = time.Now()
					w.heartbeatTimeMapLock.Unlock()
				}
			case watch.Modified:
				// update heartbeat time
				newNode := (event.Object).(*core.Node)
				if !w.isMaster(newNode) {
					w.heartbeatTimeMapLock.Lock()
					w.lastHeartbeatTimeMap[newNode.UID] = time.Now()
					w.heartbeatTimeMapLock.Unlock()
				}
			case watch.Deleted:
				// delete node from map
				oldNode := (event.Object).(*core.Node)
				nodeUID := oldNode.GetUID()
				if !w.isMaster(oldNode) {
					w.heartbeatTimeMapLock.Lock()
					delete(w.lastHeartbeatTimeMap, nodeUID)
					w.heartbeatTimeMapLock.Unlock()
				}
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

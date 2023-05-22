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

	hbCli, _ := apiclient.NewRESTClient(types.HeartbeatObjectType)
	hbListWatcher := listwatch.NewListWatchFromClient(hbCli)

	return &watcher{
		nodeClient:           nodeCli,
		nodeListWatcher:      nodeListWatcher,
		heartbeatClient:      hbCli,
		heartbeatListWatcher: hbListWatcher,
		lastHeartbeatMap:     nil,
	}
}

type watcher struct {
	nodeClient      client.Interface
	nodeListWatcher listwatch.ListerWatcher

	heartbeatClient      client.Interface
	heartbeatListWatcher listwatch.ListerWatcher

	// map of nodeUID and last heartbeat
	lastHeartbeatMap map[types.UID]core.Heartbeat // master should not be record in this map
	heartbeatMapLock sync.RWMutex
}

func (w *watcher) Run(ctx context.Context, cancel context.CancelFunc) {
	log.Printf("[HeartbeatWatcher] start\n")
	defer log.Printf("[HeartbeatWatcher] running\n")

	syncChan := make(chan bool)

	go func() {
		defer cancel()
		defer log.Printf("[HeartbeatWatcher] listAndWatchNodes finished\n")
		err := w.listAndWatchNodes(syncChan, ctx.Done())
		if err != nil {
			log.Printf("[HeartbeatWatcher] listAndWatchNodes failed, err: %v\n", err)
		}
	}()

	// wait for node list finish
	<-syncChan

	// watch heartbeat
	go func() {
		defer cancel()
		defer log.Printf("[HeartbeatWatcher] watchHeartbeats finished\n")
		err := w.watchHeartbeats(ctx.Done())
		if err != nil {
			log.Printf("[HeartbeatWatcher] watchHeartbeats failed, err: %v\n", err)
		}
	}()

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

			log.Printf("[HeartbeatWatcher] checkHeartbeat: start\n")

			// check heartbeat time
			w.heartbeatMapLock.RLock()
			for uid, hb := range w.lastHeartbeatMap {
				if hb.Spec.NodeUID == "" {
					continue
				}
				if time.Since(hb.Status.Timestamp) > config.HeartbeatDeadInterval {
					log.Printf("[HeartbeatWatcher] checkHeartbeat: node %v dead, last heartbeat time %v, current time %v\n", uid, hb.Status.Timestamp, time.Now())
					// node dead, delete that node
					_, _, err := w.nodeClient.Delete(uid)
					if err != nil {
						log.Printf("[HeartbeatWatcher] checkHeartbeat: node %v delete failed\n", uid)
					}
				}
			}
			w.heartbeatMapLock.RUnlock()

			time.Sleep(defaultWorkerSleepInterval)
		}
	}

}

func (w *watcher) watchHeartbeats(stopCh <-chan struct{}) error {

	// start watch hb change
	wi, err := w.heartbeatListWatcher.Watch()
	if err != nil {
		return err
	}

	err = w.handleWatchHeartbeats(wi, stopCh)
	wi.Stop() // stop watch

	if err == errorStopRequested {
		return nil
	}

	return err

}

func (w *watcher) handleWatchHeartbeats(wi watch.Interface, stopCh <-chan struct{}) error {
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
			// log.Printf("[handleWatchHeartbeats] event %v\n", event)
			// log.Printf("[handleWatchHeartbeats] event object %v\n", event.Object)
			eventCount += 1

			switch event.Type {
			case watch.Added, watch.Modified:
				// update heartbeat
				newHb := (event.Object).(*core.Heartbeat)
				w.heartbeatMapLock.Lock()
				w.lastHeartbeatMap[newHb.Spec.NodeUID] = *newHb
				w.heartbeatMapLock.Unlock()
			case watch.Deleted:
				// ignore
			case watch.Bookmark:
				panic("[handleWatchHeartbeats] watchHandler Event Type watch.Bookmark received")
			case watch.Error:
				panic("[handleWatchHeartbeats] watchHandler Event Type watch.Error received")
			default:
				panic("[handleWatchHeartbeats] watchHandler Unknown Event Type received")
			}
		}
	}
	return nil
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

	w.heartbeatMapLock.Lock()

	// create map (without master)
	w.lastHeartbeatMap = make(map[types.UID]core.Heartbeat, len(nodeItems)-1)

	w.heartbeatMapLock.Unlock()

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
			log.Printf("[handleWatchNodes] event %v\n", event)
			log.Printf("[handleWatchNodes] event object %v\n", event.Object)
			eventCount += 1

			switch event.Type {
			case watch.Added, watch.Modified:
				// ignore
			case watch.Deleted:
				// delete node from map
				oldNode := (event.Object).(*core.Node)
				nodeUID := oldNode.GetUID()
				if !w.isMaster(oldNode) {

					w.heartbeatMapLock.Lock()
					hb := w.lastHeartbeatMap[nodeUID]
					delete(w.lastHeartbeatMap, nodeUID)
					w.heartbeatMapLock.Unlock()

					// delete heartbeat
					_, _, err := w.heartbeatClient.Delete(hb.UID)
					if err != nil {
						log.Printf("[handleWatchNodes] node deleted notified, delete heartbeat failed\n")
					}

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

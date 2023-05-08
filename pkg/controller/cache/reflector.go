package cache

import (
	"errors"
	"log"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/types"
	"minik8s/pkg/api/watch"
	"minik8s/pkg/apiclient/listwatch"
	"time"
)

// Reflector watches a specified resource and causes all changes to be reflected in the given store.
type Reflector struct {
	// name identifies this reflector. By default it will be a file:line if possible.
	name string
	// expectedType of object of the type we expect to place in the store.
	expectedType types.ApiObjectType
	// The destination to sync up with the watch source
	store ThreadSafeStore
	// listerWatcher is used to perform lists and watches.
	listerWatcher listwatch.ListerWatcher

	// TODO implement resync
	resyncPeriod time.Duration

	// transportQueue is used to tell informer new events happening
	transportQueue WorkQueue

	//// lastSyncResourceVersion is the resource version token last
	//// observed when doing a sync with the underlying store
	//// it is thread safe, but not synchronized with the underlying store
	//lastSyncResourceVersion string
	//// isLastSyncResourceVersionUnavailable is true if the previous list or watch request with
	//// lastSyncResourceVersion failed with an "expired" or "too large resource version" error.
	//isLastSyncResourceVersionUnavailable bool
	//// lastSyncResourceVersionMutex guards read/write access to lastSyncResourceVersion
	//lastSyncResourceVersionMutex sync.RWMutex
}

// NewReflector creates a new Reflector
func NewReflector(lw listwatch.ListerWatcher, ty types.ApiObjectType, resyncPeriod time.Duration, s ThreadSafeStore, q WorkQueue) *Reflector {
	return &Reflector{
		name:           string(ty) + " Reflector",
		resyncPeriod:   resyncPeriod,
		listerWatcher:  lw,
		store:          s,
		expectedType:   ty,
		transportQueue: q,
	}
}

var (
	// Used to indicate that watching stopped because of a signal from the stop
	// channel passed in from a client of the reflector.
	errorStopRequested = errors.New("stop requested")
)

// Run uses the reflector's ListAndWatch to fetch all the
// objects and subsequent deltas.
// Run will exit when stopCh is closed.
func (r *Reflector) Run(stopCh <-chan struct{}, syncChan chan bool) error {
	log.Printf("[Reflector] Starting reflector %s (%s) from %s\n", r.expectedType, r.resyncPeriod, r.name)
	if err := r.ListAndWatch(stopCh, syncChan); err != nil {
		log.Printf("[Reflector] ListAndWatch error %v, %s (%s) from %s\n", err, r.expectedType, r.resyncPeriod, r.name)
		return err
	}
	log.Printf("[Reflector] Stopping reflector %s (%s) from %s\n", r.expectedType, r.resyncPeriod, r.name)
	return nil
}

// ListAndWatch first lists all items and get the resource version at the moment of call,
// and then use the resource version to watch.
// It returns error if ListAndWatch didn't even try to initialize watch.
func (r *Reflector) ListAndWatch(stopCh <-chan struct{}, syncChan chan bool) error {
	log.Printf("[Reflector] Listing and watching %v from %s\n", r.expectedType, r.name)

	var err error
	var w watch.Interface
	var l core.IApiObjectList

	l, err = r.listerWatcher.List()
	if err != nil {
		return err
	}

	err = r.listHandler(l)
	if err != nil {
		return err
	}

	// send signal through syncChan to tell informer list finish
	syncChan <- true

	w, err = r.listerWatcher.Watch()
	if err != nil {
		return err
	}

	err = r.watchHandler(w, stopCh)
	w.Stop() // stop watch

	if err == errorStopRequested {
		return nil
	}

	return err

}

// NotificationEvent is event put in transportQueue to tell informer
// what new event is happening, and corresponding key
type NotificationEvent struct {
	ObjKey string
	Type   watch.EventType
	Event  watch.Event
}

// pushNotificationEvent transform watch.Event to NotificationEvent, and push it into transportQueue
func (r *Reflector) pushNotificationEvent(watchEvent watch.Event) {

	ne := NotificationEvent{
		ObjKey: r.getObjectKey(watchEvent.Object),
		Type:   watchEvent.Type,
		Event:  watchEvent,
	}

	r.transportQueue.Enqueue(ne)
}

// listHandler lists l
func (r *Reflector) listHandler(l core.IApiObjectList) error {
	items := l.GetIApiObjectArr()
	for _, obj := range items {
		r.store.Add(r.getObjectKey(obj), obj)
	}
	return nil
}

// watchHandler watches w
func (r *Reflector) watchHandler(w watch.Interface, stopCh <-chan struct{}) error {
	eventCount := 0
loop:
	for {
		select {
		case <-stopCh:
			log.Printf("[Reflector] watchHandler stop received from stopCh\n")
			log.Printf("[Reflector] %s: Watch close - %v total %v items received\n", r.name, r.expectedType, eventCount)
			return errorStopRequested
		case event, ok := <-w.ResultChan():
			if !ok {
				break loop
			}
			log.Printf("[Reflector] watchHandler event %v\n", event)
			log.Printf("[Reflector] watchHandler event object %v\n", event.Object)
			eventCount += 1

			switch event.Type {
			case watch.Added, watch.Modified, watch.Deleted:
				// push NotificationEvent to queue to notify informer about new event
				r.pushNotificationEvent(event)

			case watch.Bookmark:
				panic("[Reflector] watchHandler Event Type watch.Bookmark received")
			case watch.Error:
				log.Printf("[Reflector] watchHandler watch.Error event object received %v\n", event.Object)
				log.Printf("[Reflector] %s: Watch close - %v total %v items received\n", r.name, r.expectedType, eventCount)
				return event.Object.(*core.ErrorApiObject).GetError()
			default:
				panic("[Reflector] watchHandler Unknown Event Type received")
			}
		}
	}
	log.Printf("[Reflector] %s: Watch close - %v total %v items received\n", r.name, r.expectedType, eventCount)
	return nil
}

// getObjectKey get the key of object for storing
func (r *Reflector) getObjectKey(obj core.IApiObject) string {
	name := obj.GetUID()
	return name
	// prefix := core.GetApiObjectsURL(r.expectedType)
	// return prefix + name
}

package cache

import (
	"errors"
	"log"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/watch"
	"minik8s/pkg/client/listwatch"
	"sync"
	"time"
)

// Reflector watches a specified resource and causes all changes to be reflected in the given store.
type Reflector struct {
	// name identifies this reflector. By default it will be a file:line if possible.
	name string
	// expectedType of object of the type we expect to place in the store.
	expectedType core.ApiObjectType
	// The destination to sync up with the watch source
	store ThreadSafeStore
	// listerWatcher is used to perform lists and watches.
	listerWatcher listwatch.ListerWatcher
	resyncPeriod  time.Duration

	// lastSyncResourceVersion is the resource version token last
	// observed when doing a sync with the underlying store
	// it is thread safe, but not synchronized with the underlying store
	lastSyncResourceVersion string
	// isLastSyncResourceVersionUnavailable is true if the previous list or watch request with
	// lastSyncResourceVersion failed with an "expired" or "too large resource version" error.
	isLastSyncResourceVersionUnavailable bool
	// lastSyncResourceVersionMutex guards read/write access to lastSyncResourceVersion
	lastSyncResourceVersionMutex sync.RWMutex
}

// NewReflector creates a new Reflector
func NewReflector(lw listwatch.ListerWatcher, ty core.ApiObjectType, resyncPeriod time.Duration) *Reflector {
	return &Reflector{
		name:          string(ty) + " Reflector",
		resyncPeriod:  resyncPeriod,
		listerWatcher: lw,
		store:         NewThreadSafeStore(),
		expectedType:  ty,
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
func (r *Reflector) Run(stopCh <-chan struct{}) error {
	log.Printf("[Reflector] Starting reflector %s (%s) from %s\n", r.expectedType, r.resyncPeriod, r.name)
	if err := r.ListAndWatch(stopCh); err != nil {
		log.Printf("[Reflector] ListAndWatch error %v, %s (%s) from %s\n", err, r.expectedType, r.resyncPeriod, r.name)
		return err
	}
	log.Printf("[Reflector] Stopping reflector %s (%s) from %s\n", r.expectedType, r.resyncPeriod, r.name)
	return nil
}

// ListAndWatch first lists all items and get the resource version at the moment of call,
// and then use the resource version to watch.
// It returns error if ListAndWatch didn't even try to initialize watch.
func (r *Reflector) ListAndWatch(stopCh <-chan struct{}) error {
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

// listHandler lists l
func (r *Reflector) listHandler(l core.IApiObjectList) error {
	// l.GetItems()
	// TODO: store list in store
	// 	how to get keys?
	return nil
}

// watchHandler watches w
func (r *Reflector) watchHandler(w watch.Interface, stopCh <-chan struct{}) error {
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
			log.Printf("[Reflector] watchHandler event %v\n", event)
			log.Printf("[Reflector] watchHandler event object %v\n", event.Object)
			eventCount += 1
			switch event.Type {
			case watch.Added:
				r.store.Add(event.Key, event.Object)
			case watch.Modified:
				r.store.Update(event.Key, event.Object)
			case watch.Deleted:
				r.store.Delete(event.Key)
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

package cache

import (
	"log"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/watch"
	"minik8s/pkg/client/listwatch"
	"time"
)

type Informer interface {
	AddEventHandler(handler ResourceEventHandler) error

	// Run starts and runs the informer, returning after it stops.
	// The informer will be stopped when stopCh is closed.
	Run(stopCh <-chan struct{})
}

type informer struct {
	objType   core.ApiObjectType
	reflector *Reflector
	handlers  []ResourceEventHandler

	// store is shared with reflector, which
	// stores objects of objType get from ApiServer
	store ThreadSafeStore

	// transportQueue is used to get notification from
	// reflector about new events happening
	transportQueue WorkQueue
}

// NewInformer returns a Store and a controller for populating the store
// while also providing event notifications. You should only used the returned
// Store for Get/List operations; Add/Modify/Deletes will cause the event
// notifications to be faulty.
//
// Parameters:
//   - lw is list and watch functions for the source of the resource you want to
//     be informed of.
//   - objType is an object of the type that you expect to receive.
//   - resyncPeriod: if non-zero, will re-list this often (you will get OnUpdate
//     calls, even if nothing changed). Otherwise, re-list will be delayed as
//     long as possible (until the upstream source closes the watch or times out,
//     or you stop the controller).
//   - h is the object you want notifications sent to.
func NewInformer(lw listwatch.ListerWatcher, objType core.ApiObjectType, resyncPeriod time.Duration, h ResourceEventHandler) Informer {
	s := NewThreadSafeStore()
	q := NewWorkQueue()
	return &informer{
		objType:        objType,
		reflector:      NewReflector(lw, objType, resyncPeriod, s, q),
		store:          s,
		handlers:       []ResourceEventHandler{h},
		transportQueue: q,
	}
}

func (i *informer) AddEventHandler(handler ResourceEventHandler) error {
	i.handlers = append(i.handlers, handler)
	return nil
}

func (i *informer) Run(stopCh <-chan struct{}) {

	syncChan := make(chan bool)

	// start go routine
	go func() {
		// Run reflector to start list and watch
		err := i.reflector.Run(stopCh, syncChan)
		if err != nil {
			log.Printf("[Informer] reflector run error: %v\n", err)
			return
		}
	}()

	// wait for reflector list finish
	<-syncChan

	// handle notification event from transportQueue
	for {
		select {
		case <-stopCh:
			return
		default:
			{
				if i.transportQueue.Empty() {
					continue
				}
				item, has := i.transportQueue.Dequeue()
				if !has {
					continue
				}
				notification, ok := item.(NotificationEvent)
				if !ok {
					panic("[Informer] Error, transportQueue element not NotificationEvent\n")
				}
				key := notification.ObjKey
				obj := notification.Event.Object
				switch notification.Type {
				case watch.Added, watch.Modified:

					oldObj, exist := i.store.Get(key)
					i.store.Update(key, obj)

					if exist {
						for _, handler := range i.handlers {
							handler.OnUpdate(oldObj, obj)
						}
					} else {
						for _, handler := range i.handlers {
							handler.OnAdd(obj)
						}
					}
				case watch.Deleted:
					obj, exist := i.store.Get(key)

					if exist {
						for _, handler := range i.handlers {
							handler.OnDelete(obj)
						}

						i.store.Delete(key)
					}
				default:

				}
			}
		}

	}
}

// ResourceEventHandler can handle notifications for events that
// happen to a resource. The events are informational only, so you
// can't return an error.  The handlers MUST NOT modify the objects
// received; this concerns not only the top level of structure but all
// the data structures reachable from it.
//   - OnAdd is called when an object is added.
//   - OnUpdate is called when an object is modified. Note that oldObj is the
//     last known state of the object-- it is possible that several changes
//     were combined together, so you can't use this to see every single
//     change. OnUpdate is also called when a re-list happens, and it will
//     get called even if nothing changed. This is useful for periodically
//     evaluating or syncing something.
//   - OnDelete will get the final state of the item if it is known, otherwise
//     it will get an object of type DeletedFinalStateUnknown. This can
//     happen if the watch is closed and misses the delete event and we don't
//     notice the deletion until the subsequent re-list.
type ResourceEventHandler interface {
	OnAdd(obj interface{})
	OnUpdate(oldObj, newObj interface{})
	OnDelete(obj interface{})
}

// ResourceEventHandlerFuncs is an implementation of ResourceEventHandler
type ResourceEventHandlerFuncs struct {
	AddFunc    func(obj interface{})
	UpdateFunc func(oldObj, newObj interface{})
	DeleteFunc func(obj interface{})
}

// OnAdd calls AddFunc if it's not nil.
func (r ResourceEventHandlerFuncs) OnAdd(obj interface{}) {
	if r.AddFunc != nil {
		r.AddFunc(obj)
	}
}

// OnUpdate calls UpdateFunc if it's not nil.
func (r ResourceEventHandlerFuncs) OnUpdate(oldObj, newObj interface{}) {
	if r.UpdateFunc != nil {
		r.UpdateFunc(oldObj, newObj)
	}
}

// OnDelete calls DeleteFunc if it's not nil.
func (r ResourceEventHandlerFuncs) OnDelete(obj interface{}) {
	if r.DeleteFunc != nil {
		r.DeleteFunc(obj)
	}
}

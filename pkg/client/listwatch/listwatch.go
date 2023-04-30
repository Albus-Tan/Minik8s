package listwatch

import (
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/watch"
)

// Lister is any object that knows how to perform an initial list.
type Lister interface {
	// List should return a list type object; the Items field will be extracted, and the
	// ResourceVersion field will be used to start the watch in the right place.
	List(options meta.ListOptions) ([]core.IApiObject, error)
}

// Watcher is any object that knows how to start a watch on a resource.
type Watcher interface {
	// Watch should begin a watch at the specified version.
	Watch(options meta.ListOptions) (watch.Interface, error)
}

// ListerWatcher is any object that knows how to perform an initial list and start a watch on a resource.
type ListerWatcher interface {
	Lister
	Watcher
}

// ListFunc knows how to list resources
type ListFunc func(options meta.ListOptions) ([]core.IApiObject, error)

// WatchFunc knows how to watch resources
type WatchFunc func(options meta.ListOptions) (watch.Interface, error)

// ListWatch knows how to list and watch a set of apiserver resources.  It satisfies the ListerWatcher interface.
// It is a convenience function for users of NewReflector, etc.
// ListFunc and WatchFunc must not be nil
type ListWatch struct {
	ListFunc  ListFunc
	WatchFunc WatchFunc
	// DisableChunking requests no chunking for this list watcher.
	DisableChunking bool
}

// List a set of apiserver resources
func (lw *ListWatch) List(options meta.ListOptions) ([]core.IApiObject, error) {
	// ListWatch is used in Reflector, which already supports pagination.
	// Don't paginate here to avoid duplication.
	return lw.ListFunc(options)
}

// Watch a set of apiserver resources
func (lw *ListWatch) Watch(options meta.ListOptions) (watch.Interface, error) {
	return lw.WatchFunc(options)
}

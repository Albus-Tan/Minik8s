package listwatch

import (
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/watch"
	client "minik8s/pkg/client/interface"
)

// Lister is any object that knows how to perform an initial list.
type Lister interface {
	// List should return a list type object; the Items field will be extracted, and the
	// ResourceVersion field will be used to start the watch in the right place.
	List() (core.IApiObjectList, error)
}

// Watcher is any object that knows how to start a watch on a resource.
type Watcher interface {
	// Watch should begin a watch at the specified version.
	Watch() (watch.Interface, error)
}

// ListerWatcher is any object that knows how to perform an initial list and start a watch on a resource.
type ListerWatcher interface {
	Lister
	Watcher
}

// ListFunc knows how to list resources
type ListFunc func(options meta.ListOptions) (core.IApiObjectList, error)

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
func (lw *ListWatch) List() (core.IApiObjectList, error) {
	// ListWatch is used in Reflector, which already supports pagination.
	// Don't paginate here to avoid duplication.
	var options meta.ListOptions
	return lw.ListFunc(options)
}

// Watch a set of apiserver resources
func (lw *ListWatch) Watch() (watch.Interface, error) {
	var options meta.ListOptions
	return lw.WatchFunc(options)
}

// NewListWatchFromClient creates a new ListWatch from the specified client, resource, namespace and field selector.
func NewListWatchFromClient(c client.Interface) *ListWatch {
	optionsModifier := func(options *meta.ListOptions) {
		// options.FieldSelector = fieldSelector.String()
	}
	return NewFilteredListWatchFromClient(c, optionsModifier)
}

// NewFilteredListWatchFromClient creates a new ListWatch from the specified client, resource, namespace, and option modifier.
// Option modifier is a function takes a ListOptions and modifies the consumed ListOptions. Provide customized modifier function
// to apply modification to ListOptions with a field selector, a label selector, or any other desired options.
func NewFilteredListWatchFromClient(c client.Interface, optionsModifier func(options *meta.ListOptions)) *ListWatch {
	listFunc := func(options meta.ListOptions) (core.IApiObjectList, error) {
		optionsModifier(&options)
		return c.GetAll()
	}
	watchFunc := func(options meta.ListOptions) (watch.Interface, error) {
		options.Watch = true
		optionsModifier(&options)
		return c.WatchAll()
	}
	return &ListWatch{ListFunc: listFunc, WatchFunc: watchFunc}
}

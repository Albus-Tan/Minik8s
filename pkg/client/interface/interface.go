package client

import (
	"minik8s/pkg/api"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/watch"
)

// Interface captures the set of operations for generically interacting with Kubernetes REST apis.
type Interface interface {
	Post(object core.IApiObject) (int, *api.PostResponse, error)
	Put(name string, object core.IApiObject) (int, *api.PutResponse, error)
	Get(name string) (core.IApiObject, error)
	GetStatus(name string) (core.IApiObjectStatus, error)
	PutStatus(name string, object core.IApiObjectStatus) (int, *api.PutResponse, error)
	GetAll() (objectList core.IApiObjectList, err error)
	Delete(name string) (string, error)
	WatchAll() (watch.Interface, error)
	Watch(name string) (watch.Interface, error)
	URL() string
	WatchURL() string
}

# API

> API 标准参照 https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.22/

- API 接口规定为 [BindHandlers](../pkg/apiserver/httpserver.go) 函数中的注释及对应注册的 [api url](../pkg/api/url.go)
- API object 定义在 `/pkg/api` 路径下

# Concurrency Control and Consistency

> Ref：https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency

All resources have a "resourceVersion" field as part of their metadata. This resourceVersion is a string that identifies the internal version of an object that can be used by clients to determine when objects have changed. When a record is about to be updated, its version is checked against a pre-saved value, and if it doesn't match, the update fails with a StatusConflict (HTTP status code 409).

The resourceVersion is currently backed by [etcd's mod_revision](https://etcd.io/docs/latest/learning/api/#key-value-pair). However, it's important to note that the application should *not* rely on the implementation details of the versioning system maintained by Kubernetes. We may change the implementation of resourceVersion in the future, such as to change it to a timestamp or per-object counter.
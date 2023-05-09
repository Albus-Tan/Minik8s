# ApiServer

作为 server 接收其他组件增删改查与监听的请求，并将对应修改存储进 `etcd`

- `handlers`：注册的处理函数
- `httpserver`：使用 `gin` 作为服务器框架
- `etcd`：通过 `etcd client` 与 `etcd` 直接交互，进行键值对存取

## Concurrency Control and Consistency

> Ref：https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#concurrency-control-and-consistency

All resources have a "resourceVersion" field as part of their metadata. This resourceVersion is a string that identifies the internal version of an object that can be used by clients to determine when objects have changed. When a record is about to be updated, its version is checked against a pre-saved value, and if it doesn't match, the update fails with a StatusConflict (HTTP status code 409).

The resourceVersion is currently backed by [etcd's mod_revision](https://etcd.io/docs/latest/learning/api/#key-value-pair). However, it's important to note that the application should *not* rely on the implementation details of the versioning system maintained by Kubernetes. We may change the implementation of resourceVersion in the future, such as to change it to a timestamp or per-object counter.

## Watch

通过 `etcd` `Watch` 对 `key` 进行监听，每当对应 `value` 发生修改，就会通过 channel 进行通知

- watch 时 delete 的响应 value 为 `“”`，为了获取被 delete 的内容需要使用 `clientv3.WithPrevKV()`

# ApiClient

 `client.Interface` 是所有能够与 `ApiServer` 交互的 client 的统一接口，应当通过这些接口使用 client，而不要直接创建实现的实例

- `http`：包括对于发送 REST http 请求及处理响应的封装
- `RESTClient`：可以与 ApiServer 交互的 client，每种资源类型一个 client，不同资源类型不能共用！
  - 其中 watch 相关的方法通过 StreamWatcher 实现

# ListWatch

通过 `client.Interface` 创建，封装接口，专门用来调用对应资源的 `GetAll` 与 `WatchAll` 方法

## StreamWatcher

`ListWatcher` 的实现

**组件**

- `Decoder`：负责将 `Watch` 监听到的 `ApiServer` 发来的事件类型转换为 `watch.Event` 类型
  - 此处 `ApiServer` 发来的事件类型为 `Etcd` 内置事件类型，这么做的好处在于解耦，修改实现只需要实现对应 `Decoder interface` 接口即可
- `Reporter`：错误处理，将报告错误的事件转换为标准的 `watch.Event` 类型，同时也将过程中产生的错误 转换为标准的 `watch.Event` 类型
- `chan Event`：通过此通道将监听到的事件发送出去，使用 `StreamWatcher` 的组件可以通过 `watch.Interface` 中的 `ResultChan()` 拿到这个通道，并获取事件；通道使用完成后/结束时需要通过 `Stop()` 方法关闭通道


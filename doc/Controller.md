# Controller

## Informer

相当于每个 Node 上对于不同 ApiObject 的本地缓存，每一种 ApiObject 资源对应一个 Informer（由 `objType` 指定）

- 启动时其中的 `Reflector` 通过 `List` 向 ApiServer 拿取所有 ApiObject 信息，并且存储在 `ThreadSafeStore` 中
- 之后其 `Reflector` 通过 `Watch` 监听所有对应 ApiObject 的变化事件，并存储在 `ThreadSafeStore` 中，同时调用注册进来的 `ResourceEventHandler` 进行相应处理

**组件**

- `Reflector`：启动时先通过 `List` 向 ApiServer 拿取所有 ApiObject 信息，并且存储在 `ThreadSafeStore` 中，之后通过 `Watch` 监听所有对应 ApiObject 的变化事件，并通知  `Informer` （通过将事件放入 `WorkQueue`）；其中 `List` 与 `Watch` 都由 `listwatch.ListerWatcher` 组件完成
- `ThreadSafeStore`：与其 `Reflector` 共享同一个存储，存储对应 ApiObject 对象的本地缓存
- `ResourceEventHandler`：注册对于各种 `Watch` 事件的响应，使用 `Informer` 的组件可以通过 `AddEventHandler ` 添加对应处理函数
- `WorkQueue`：每次 `Reflector` 监听到新事件，就放进此队列，等待 `Informer` 在 `run` 中进行处理，并调用相应注册进来的 `EventHandler` 函数

## WorkQueue

- 线程安全的队列，通过读写锁允许多个线程同时处理而不出现并发问题
- 在 `Dequeue` 时如果队列为空，会通过 conditional variable 等待 `Enqueue` 唤醒，再进行 `Dequeue`

# ReplicaSet Controller

维护与 `selector` `matchLabels` 标签匹配的 `replicas` 数量的 Pod，多删少增

- 注意原本受到 ReplicaSet 管理的 Pod 的 label 发生更新时，需要重新检查是否符合 ReplicaSet 的 `selector` 匹配，否的话需要新接管 Pod，并把这个不再被管理的 Pod 的对应 ReplicaSet `OwnerReference` 字段去掉
- 注意当创建 ReplicaSet 时，如果已经有 Pod，并且其 `label` 匹配 ReplicaSet 的 `selector`，直接接管这些 Pod；此后没有这样满足要求的 Pod 才根据模板 `template` 创建新的 Pod


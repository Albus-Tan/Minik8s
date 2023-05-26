# Node

不同物理实体机对应的 Node config 文件的 Name 字段需要保证 unique

## Heartbeat

所有 worker 节点启动后，会由 Heartbeat Sender 持续向 Master 节点的 Heartbeat Watcher 发送心跳，一旦  Heartbeat Watcher 一段时间没有接收到 worker 节点发来的心跳，就认为对应 worker 节点挂掉，并将其信息在 etcd 内删除

## Init

Node 初始化时会检查当前有无 Node 与其重名，如果有，判断 config 文件是否与已有 Node 信息不同

- 如果一致，则复用当前 Node，不再创建新 Node
- 如果不一致，报错给用户并退出；用户需要修改 config 文件的 Name 字段，或通过 put 方式修改原有 Node 的配置文件相关内容，以实现配置的修改
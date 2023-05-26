# Scheduler

调度器共支持三种调度策略，分别为 `NodeAffinity`，`PodAntiAffinity` 与 `Round Robin`

-  `NodeAffinity`：Pod 可以直接指定希望在哪个 Node 上运行（通过在 yaml 配置文件中指定 Node name）
- `PodAntiAffinity`：Pod 可以指定和拥有某种 label 的 Pod 不运行在相同的 Node 上；调度时会尽可能满足 Pod 的 AntiAffinity 需求，当然如果当前所有 Node 都不能满足（比如所有 Node 上都跑了所指定的不能与其一同运行的 Pod），则此配置不生效
-  `Round Robin`：新来的 Pod 依次轮流调度到各个 Node 上；期间通过 `NodeAffinity` 调度的 Pod 不会影响 RR 队列，通过 `PodAntiAffinity` 调度的 Pod 会将被调度到的节点置于 RR 队列的末尾

## 调度逻辑

1. 在调度时，会首先判断新创建的 Pod 有无指定 `NodeAffinity`（通过 Pod 的 Spec 中 Node name 字段），如果有则直接调度至对应 node，无则判断有无指定 `PodAntiAffinity`
2. 如果指定了 `PodAntiAffinity`，会尝试采用此策略进行调度；否则直接采用默认的 `Round Robin` 策略调度
   -  `PodAntiAffinity` 中会通过新创建 Pod 的 label selector 判断各个 Node 上现有的 Pod 的 label 是否与其相符，来决定新 Pod 不能调度到哪些 Node 上；如果所有 Node 都被排除，会无视反亲和性配置，采用  `Round Robin` 进行调度
   - 如果通过 `PodAntiAffinity` 调度成功，会将 RR 队列中对应的 Node 移到末尾
3.  `Round Robin` 策略通过维护一个 Node 队列实现，每次调度时取队首 Node ，之后将对应 Node 放置队尾，实现 RR 目的
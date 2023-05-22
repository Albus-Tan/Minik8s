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

# Autoscaling Controller

- `runWorker`：从工作队列中拿出对应 hpa，并检查是否满足扩缩容条件，进行自动扩缩容
- `periodicallyCheckScale`：周期性地将所有 hpa 放入工作队列中（每 15 秒），以实现周期性检查，来达到自动扩缩容的目的

## Cadvisor 资源指标监控

**使用二进制部署**

```sh
# 下载二进制
https://github.com/google/cadvisor/releases/latest
# 本地运行
./cadvisor  -port=8090 &>>/var/log/cadvisor.log
# 查看进程信息
ps -aux | grep cadvisor
# 查看端口占用
netstat -anp | grep 8090
```

**使用docker部署**

```bash
docker run \
--volume=/:/rootfs:ro \
--volume=/var/run:/var/run:rw \
--volume=/sys:/sys:ro \
--volume=/var/lib/docker/:/var/lib/docker:ro \
--volume=/dev/disk/:/dev/disk:ro \
--publish=8090:8090 \
--detach=true \
--name=cadvisor \
google/cadvisor:latest
```

**端口转发**

这样在本机上就可以看到远端机器上的 cadvisor

```
ssh -N minik8s-dev -L 8090:localhost:8090
```

## ResourceMetricsClient

- 通过一系列 `cadvisorClients map[types.UID]cadvisor.Interface` （每个 node 一个 cadvisor client），来获取每个 node 的各项资源占用指标，并进行聚合
- 每隔一段时间同步更新最新 node 信息，以及资源指标信息

## HorizontalPodAutoscaler

### HorizontalPodAutoscalerSpec

- `ScaleTargetRef`：HPA 控制的对象的 ObjectReference（当前仅支持 ReplicaSet）
- `MinReplicas`：HPA 自动扩缩时最少能缩到多少 Pod
- `MaxReplicas`：HPA 自动扩缩时最多能扩到多少 Pod
- `MetricSpec`：扩缩容决策所依据的资源指标
- `HorizontalPodAutoscalerBehavior`：扩缩容策略

#### MetricSpec 扩缩容决策所依据的资源指标

定义了在当前资源指标的量化标准下，应该怎么 scale

- `MetricSourceType`：资源指标类型，目前仅支持 `ResourceMetricSourceType`
- `ResourceMetricSource`：当资源指标类型为 `ResourceMetricSourceType` 时不为空

```go
type MetricSpec struct {
	Type MetricSourceType `json:"type"`
	Resource *ResourceMetricSource `json:"resource,omitempty"`
}
```

`ResourceMetricSource` 包括如下字段：

- `ResourceName`：资源名称，目前支持 `ResourceCPU` 和 `ResourceMemory`
- `MetricTarget`：对应 `ResourceName` 资源的目标值

```go
type ResourceMetricSource struct {
	Name types.ResourceName `json:"name"`
	Target MetricTarget `json:"target"`
}
```

`MetricTarget` 包括如下字段：

- `MetricTargetType`：可以是 `Value`, `AverageValue` 或 `Utilization`（目前支持 `AverageValue` 和 `Utilization`）
- `Value`：指标的目标值
- `AverageValue`：指标的目标值（对于所有相关 pod 计算该指标平均值）
- `AverageUtilization`：指标的目标值，百分比表示（对于所有相关 pod 计算该指标平均值）

```go
type MetricTarget struct {
   // Type represents whether the metric type is Utilization, Value, or AverageValue
   Type MetricTargetType `json:"type"`
   // Value is the target value of the metric (as a quantity).
   Value types.Quantity `json:"value,omitempty"`
   // TargetAverageValue is the target value of the average of the
   // metric across all relevant pods (as a quantity)
   AverageValue types.Quantity `json:"averageValue,omitempty"`

   // AverageUtilization is the target value of the average of the
   // resource metric across all relevant pods, represented as a percentage of
   // the requested value of the resource for the pods.
   // Currently only valid for Resource metric source type
   AverageUtilization int32 `json:"averageUtilization,omitempty"`
}
```

#### HorizontalPodAutoscalerBehavior 扩缩容策略

可以在 `HorizontalPodAutoscalerBehavior` 中分别自定义自动扩容 `ScaleUp` 与缩容 `ScaleDown` 的策略组 `HPAScalingRules`：

```go
type HorizontalPodAutoscalerBehavior struct {
	ScaleUp *HPAScalingRules `json:"scaleUp,omitempty"`
	ScaleDown *HPAScalingRules `json:"scaleDown,omitempty"`
}
```

其中 `HPAScalingRules` 包括三个字段：

- `StabilizationWindowSeconds`：从上一次 auto scale 事件开始必须经过 `StabilizationWindowSeconds` 的秒数，才可以进行下一次的自动 scale
- `SelectPolicy`：对所配置策略组 `Policies` 中各个 policy 结果如何综合（每个 policy 规范 scale 所扩/缩的 pod 数量应当不大于多少）
  - `MaxPolicySelect`：从 Policies 的所有 Policy 中选出扩/缩的 pod 数量最多的
  - `MinPolicySelect`：从 Policies 的所有 Policy 中选出扩/缩的 pod 数量最少的
  - `DisabledPolicySelect`：禁止这一维度的 scale（也即不允许自动扩容 `ScaleUp` 或自动缩容 `ScaleDown` ）
- `Policies`：具体策略（组）

```go
type HPAScalingRules struct {
	StabilizationWindowSeconds int32 `json:"stabilizationWindowSeconds,omitempty"`
	SelectPolicy ScalingPolicySelect `json:"selectPolicy,omitempty"`
	Policies []HPAScalingPolicy `json:"policies,omitempty"`
}
```

`HPAScalingPolicy` 包括如下字段：

- `HPAScalingPolicyType`：
  - `PodsScalingPolicy`：表示 scale 所对 pod 数量做的变化 delta 需要小于等于 Value 的数值（限定变化的绝对数量）
  - `PercentScalingPolicy`：此时 Value 对应 0 至 100，表示百分之几；表示 scale 所对 pod 数量做的变化 delta 需要小于等于当前现有 pod 数量的百分之多少（如 Value 为 100，则 scale 的增/减数量至多为当前 pod 数量个 pod，也即至多倍增/全删）
- `PeriodSeconds`：从上一次 auto scale 事件开始必须经过`PeriodSeconds`的秒数，此 Policy 才可以生效

```go
type HPAScalingPolicy struct {
	Type HPAScalingPolicyType `json:"type"`
	Value int32 `json:"value"`
	PeriodSeconds int32 `json:"periodSeconds"`
}
```

**默认**

如果未定义策略，则默认如下：

```go
// scaleUp is scaling policy for scaling Up.
// If not set, the default value is the higher of:
//   * increase no more than 1 pod per 15 seconds
//   * double the number of pods per 60 seconds
// No stabilization is used.
func DefaultScaleUpRule() (r *HPAScalingRules) {
	upPolicies := make([]HPAScalingPolicy, 2)
	upPolicies[0] = HPAScalingPolicy{
		Type:          PercentScalingPolicy,
		Value:         100,
		PeriodSeconds: 60,
	}
	upPolicies[1] = HPAScalingPolicy{
		Type:          PodsScalingPolicy,
		Value:         1,
		PeriodSeconds: 15,
	}
	return &HPAScalingRules{
		StabilizationWindowSeconds: 0,
		SelectPolicy:               MaxPolicySelect,
		Policies:                   upPolicies,
	}
}
```

```go
// scaleDown is scaling policy for scaling Down.
// If not set, the default value is to allow to scale down to minReplicas pods, with a
// 300 second stabilization window (i.e., the highest recommendation for
// the last 300sec is used).
func DefaultScaleDownRule() (r *HPAScalingRules) {
	downPolicies := make([]HPAScalingPolicy, 1)
	downPolicies[0] = HPAScalingPolicy{
		Type:          PercentScalingPolicy,
		Value:         100,
		PeriodSeconds: 15,
	}
	return &HPAScalingRules{
		StabilizationWindowSeconds: 300,
		SelectPolicy:               MinPolicySelect,
		Policies:                   downPolicies,
	}
}
```


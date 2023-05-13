package metrics

import (
	info "github.com/google/cadvisor/info/v1"
	"minik8s/config"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/types"
	"minik8s/pkg/apiclient"
	client "minik8s/pkg/apiclient/interface"
	"minik8s/pkg/apiclient/listwatch"
	"minik8s/pkg/cadvisor"
	"minik8s/pkg/logger"
	"sync"
	"time"
)

type resourceMetricsClient struct {
	// cadvisorClients map with node uid as key and client value
	// each node will have one cadvisor client to get container metrics
	// on it, the node info will be updated when CollectAllMetrics is
	// called and time from timeLastSynced over defaultSyncInterval
	cadvisorClients map[types.UID]cadvisor.Interface

	nodeClient      client.Interface
	nodeListWatcher listwatch.ListerWatcher

	mtx sync.Mutex

	timeLastSynced time.Time
}

func NewResourceMetricsClient() MetricsClient {

	nodeClient, _ := apiclient.NewRESTClient(types.NodeObjectType)
	lw := listwatch.NewListWatchFromClient(nodeClient)
	cadvisorClients := make(map[types.UID]cadvisor.Interface)

	rmc := &resourceMetricsClient{
		cadvisorClients: cadvisorClients,
		nodeClient:      nodeClient,
		nodeListWatcher: lw,
	}

	rmc.mtx.Lock()
	defer rmc.mtx.Unlock()

	nodeList, err := lw.List()
	if err != nil {
		logger.ControllerManagerLogger.Printf("[MetricsClient] list nodes failed when creating, err: %v\n", err)
	} else {
		nodes := nodeList.GetIApiObjectArr()
		cadvisorClients = make(map[types.UID]cadvisor.Interface, len(nodes))

		rmc.timeLastSynced = time.Now()

		logger.ControllerManagerLogger.Printf("[MetricsClient] New, list nodes length: %v\n", len(nodes))

		for _, nodeItem := range nodes {
			node := nodeItem.(*core.Node)
			cli := cadvisor.NewClient(config.CadvisorUrl(node.Spec.Address))
			cadvisorClients[node.GetUID()] = cli
		}
	}

	return rmc
}

const defaultSyncInterval = 30 * time.Second

func (r *resourceMetricsClient) syncNodeInfo() {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	if time.Since(r.timeLastSynced) < defaultSyncInterval {
		return
	}

	r.timeLastSynced = time.Now()

	nodeList, err := r.nodeListWatcher.List()

	if err != nil {
		logger.ControllerManagerLogger.Printf("[MetricsClient] list nodes filed when syncNodeInfo, err: %v\n", err)
	} else {
		nodes := nodeList.GetIApiObjectArr()

		if len(nodes) == len(r.cadvisorClients) {
			return
		}

		logger.ControllerManagerLogger.Printf("[MetricsClient] syncNodeInfo, current %v nodes\n", len(nodes))

		r.cadvisorClients = make(map[types.UID]cadvisor.Interface, len(nodes))

		for _, nodeItem := range nodes {
			node := nodeItem.(*core.Node)
			cli := cadvisor.NewClient(config.CadvisorUrl(node.Spec.Address))
			r.cadvisorClients[node.GetUID()] = cli
		}
	}

}

func ContainerKeyFunc(id string) string {
	return id
}

func AddPodMetric(a *PodMetric, b *PodMetric) PodMetric {
	return PodMetric{
		Timestamp: a.Timestamp,
		CpuUsage:  a.CpuUsage + b.CpuUsage,
		MemUsage:  a.MemUsage + b.MemUsage,
	}
}

func AverageStatsMetrics(stats []*info.ContainerStats) *PodMetric {
	var cpuUsg uint64 = 0
	var memUsg uint64 = 0
	var cnt uint64 = 0
	for _, stat := range stats {
		if stat != nil {
			cpuUsg += stat.Cpu.Usage.Total
			memUsg += stat.Memory.Usage
			cnt += 1
		}
	}
	if cnt == 0 {
		logger.ControllerManagerLogger.Printf("[MetricsClient] Warning: AverageStatsMetrics, 0 item in ContainerStats\n")
		return &PodMetric{
			Timestamp: time.Now(),
			CpuUsage:  0,
			MemUsage:  0,
		}
	} else {
		return &PodMetric{
			Timestamp: stats[0].Timestamp,
			CpuUsage:  cpuUsg / cnt,
			MemUsage:  memUsg / cnt,
		}
	}

}

func RearrangeContainerMetricsByPods(containerMetrics map[string]info.ContainerInfo, pods []core.Pod) (podMetrics PodMetricsInfo) {
	// containerKeys map, key is container key, value is pod UID it belongs
	containerKeys := make(map[string]types.UID)
	for _, pod := range pods {
		for _, container := range pod.Status.ContainerStatuses {
			key := ContainerKeyFunc(container.ContainerID)
			containerKeys[key] = pod.UID
		}
	}

	podMetrics = make(PodMetricsInfo)

	for key, podUID := range containerKeys {
		containerInfo := containerMetrics[key]
		pm := AverageStatsMetrics(containerInfo.Stats)
		oldpm, exist := podMetrics[podUID]
		if exist {
			podMetrics[podUID] = AddPodMetric(pm, &oldpm)
		} else {
			podMetrics[podUID] = *pm
		}
	}

	return podMetrics
}

func (r *resourceMetricsClient) CollectAllMetrics() (res map[string]info.ContainerInfo, err error) {

	logger.ControllerManagerLogger.Printf("[MetricsClient] CollectAllMetrics start\n")

	r.syncNodeInfo()
	query := info.DefaultContainerInfoRequest()
	res = make(map[string]info.ContainerInfo)

	r.mtx.Lock()
	defer r.mtx.Unlock()

	for uid, cli := range r.cadvisorClients {
		containerInfos, err := cli.AllDockerContainers(&query)
		if err != nil {
			return res, err
		}

		logger.ControllerManagerLogger.Printf("[MetricsClient] AllDockerContainers info collected from node uid %v\n", uid)
		logger.ControllerManagerLogger.Printf("[MetricsClient] AllDockerContainers info: %+v\n", containerInfos)

		for _, containerInfo := range containerInfos {
			res[ContainerKeyFunc(containerInfo.Id)] = containerInfo
		}
	}
	return res, nil
}

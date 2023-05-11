package metrics

import (
	"context"
	"minik8s/config"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/types"
	"minik8s/pkg/cadvisor"
	"time"
)

type resourceMetricsClient struct {
	// cadvisorClients map with node uid as key and client value
	cadvisorClients map[types.UID]cadvisor.Interface
}

func NewResourceMetricsClient(nodes []core.Node) {
	cadvisorClients := make(map[types.UID]cadvisor.Interface, len(nodes))
	for _, node := range nodes {
		cli := cadvisor.NewClient(config.CadvisorUrl(node.Spec.Address))
		cadvisorClients[node.UID] = cli
	}
}

func (r *resourceMetricsClient) GetResourceMetric(ctx context.Context, resource types.ResourceName, namespace string, container string) (PodMetricsInfo, time.Time, error) {
	// TODO

	//metrics, err := r.client.PodMetricses(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector.String()})
	//if err != nil {
	//	return nil, time.Time{}, fmt.Errorf("unable to fetch metrics from resource metrics API: %v", err)
	//}
	//
	//if len(metrics.Items) == 0 {
	//	return nil, time.Time{}, fmt.Errorf("no metrics returned from resource metrics API")
	//}
	//var res PodMetricsInfo
	//if container != "" {
	//	res, err = getContainerMetrics(metrics.Items, resource, container)
	//	if err != nil {
	//		return nil, time.Time{}, fmt.Errorf("failed to get container metrics: %v", err)
	//	}
	//} else {
	//	res = getPodMetrics(ctx, metrics.Items, resource)
	//}
	//timestamp := metrics.Items[0].Timestamp.Time
	//return res, timestamp, nil
	//TODO implement me
	panic("implement me")
}

func (r *resourceMetricsClient) GetRawMetric(metricName string, namespace string) (PodMetricsInfo, time.Time, error) {
	//TODO implement me
	panic("implement me")
}

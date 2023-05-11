package metrics

import (
	"context"
	"minik8s/pkg/api/types"
	"time"
)

// PodMetric contains pod metric value (the metric values are expected to be the metric as a milli-value)
type PodMetric struct {
	Timestamp time.Time
	Window    time.Duration
	Value     int64
}

// PodMetricsInfo contains pod metrics as a map from pod names to PodMetricsInfo
type PodMetricsInfo map[string]PodMetric

// MetricsClient knows how to query a remote interface to retrieve container-level
// resource metrics as well as pod-level arbitrary metrics
type MetricsClient interface {
	// GetResourceMetric gets the given resource metric (and an associated oldest timestamp)
	// for the specified named container in all pods matching the specified selector in the given namespace and when
	// the container is an empty string it returns the sum of all the container metrics.
	GetResourceMetric(ctx context.Context, resource types.ResourceName, namespace string, container string) (PodMetricsInfo, time.Time, error)

	// GetRawMetric gets the given metric (and an associated oldest timestamp)
	// for all pods matching the specified selector in the given namespace
	GetRawMetric(metricName string, namespace string) (PodMetricsInfo, time.Time, error)
}

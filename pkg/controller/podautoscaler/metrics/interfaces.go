package metrics

import (
	info "github.com/google/cadvisor/info/v1"
	"minik8s/pkg/api/types"
	"time"
)

// PodMetric contains pod metric value (the metric values are expected to be the metric as a milli-value)
type PodMetric struct {
	Timestamp time.Time `json:"time,omitempty"`

	// Total CPU usage in CpuStats
	// Unit: nanoseconds.
	CpuUsage uint64 `json:"cpuUsage,omitempty"`

	// Current memory usage, this includes all memory regardless of when it was
	// accessed, in MemoryStats
	// Units: Bytes.
	MemUsage uint64 `json:"memoryUsage,omitempty"`
}

// PodMetricsInfo contains pod metrics as a map from pod UID to PodMetricsInfo
type PodMetricsInfo map[types.UID]PodMetric

// MetricsClient knows how to query a remote interface to retrieve container-level
// resource metrics as well as pod-level arbitrary metrics
type MetricsClient interface {
	CollectAllMetrics() (res map[string]info.ContainerInfo, err error)
}

package metrics

import (
	"fmt"
	"minik8s/pkg/api/types"
	"minik8s/pkg/logger"
)

// GetResourceUtilizationRatio takes in a set of metrics, a set of matching requests,
// and a target utilization percentage, and calculates the ratio of
// desired to actual utilization (returning that, the actual utilization, and the raw average value)
func GetResourceUtilizationRatio(name types.ResourceName, metrics PodMetricsInfo, requests map[types.UID]uint64, targetUtilization int32) (utilizationRatio float64, currentUtilization int32, rawAverageValue uint64, err error) {
	metricsTotal := uint64(0)
	requestsTotal := uint64(0)
	numEntries := 0

	for podUID, metric := range metrics {
		request, hasRequest := requests[podUID]
		if !hasRequest {
			// we check for missing requests elsewhere, so assuming missing requests == extraneous metrics
			continue
		}

		switch name {
		case types.ResourceMemory:
			metricsTotal += metric.MemUsage / (1024 * 1024)
		case types.ResourceCPU:
			metricsTotal += metric.CpuUsage / 1000000
		default:
			continue
		}

		requestsTotal += request
		numEntries++
	}

	// if the set of requests is completely disjoint from the set of metrics,
	// then we could have an issue where the requests total is zero
	if requestsTotal == 0 {
		return 0, 0, 0, fmt.Errorf("no metrics returned matched known pods")
	}

	currentUtilization = int32((metricsTotal) * 100 / requestsTotal)

	logger.ControllerManagerLogger.Printf("[Metrics GetResourceUtilizationRatio] numEntries %v, requestsTotal %v, metricsTotal%v, currentUtilization %v\n", numEntries, requestsTotal, metricsTotal, currentUtilization)

	return float64(currentUtilization) / float64(targetUtilization), currentUtilization, metricsTotal / uint64(numEntries), nil
}

// GetMetricUsageRatio takes in a set of metrics and a target usage value,
// and calculates the ratio of desired to actual usage
// (returning that and the actual usage)
func GetMetricUsageRatio(name types.ResourceName, metrics PodMetricsInfo, targetUsage uint64) (usageRatio float64, currentUsage uint64) {
	metricsTotal := uint64(0)
	for _, metric := range metrics {
		switch name {
		case types.ResourceCPU:
			metricsTotal += metric.CpuUsage / 1000000
		case types.ResourceMemory:
			metricsTotal += metric.MemUsage / (1024 * 1024)
		default:
			continue
		}
	}
	currentUsage = metricsTotal / uint64(len(metrics))

	return float64(currentUsage) / float64(targetUsage), currentUsage
}

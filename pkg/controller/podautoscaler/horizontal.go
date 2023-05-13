package podautoscaler

import (
	"context"
	"errors"
	"fmt"
	"math"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/types"
	client "minik8s/pkg/apiclient/interface"
	"minik8s/pkg/controller/cache"
	"minik8s/pkg/controller/podautoscaler/metrics"
	"minik8s/pkg/logger"
	"time"
)

// HorizontalController is responsible for the synchronizing HPA objects stored
// in the system with the actual deployments/replication controllers they
// control.
type HorizontalController interface {
	Run(ctx context.Context)
}

func NewHorizontalController(podInformer cache.Informer, podClient client.Interface, hpaInformer cache.Informer, hpaClient client.Interface, rsInformer cache.Informer, rsClient client.Interface) HorizontalController {

	hc := &horizontalController{
		podClient:     podClient,
		hpaClient:     hpaClient,
		rsClient:      rsClient,
		podInformer:   podInformer,
		hpaInformer:   hpaInformer,
		rsInformer:    rsInformer,
		queue:         cache.NewWorkQueue(),
		Kind:          string(types.HorizontalPodAutoscalerObjectType),
		metricsClient: metrics.NewResourceMetricsClient(),
	}

	_ = hc.hpaInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    hc.addHpa,
		UpdateFunc: hc.updateHpa,
		DeleteFunc: hc.deleteHpa,
	})

	return hc
}

type horizontalController struct {
	metricsClient metrics.MetricsClient

	hpaInformer cache.Informer
	hpaClient   client.Interface

	podInformer cache.Informer
	podClient   client.Interface

	rsInformer cache.Informer
	rsClient   client.Interface

	queue cache.WorkQueue
	Kind  string
}

func (h *horizontalController) Run(ctx context.Context) {

	go func() {
		logger.HorizontalControllerLogger.Printf("[HorizontalController] start\n")
		defer logger.HorizontalControllerLogger.Printf("[HorizontalController] finish\n")

		h.runWorker(ctx)

		h.periodicallyCheckScale()

		// wait for controller manager stop
		<-ctx.Done()
	}()
	return
}

func (h *horizontalController) HpaKeyFunc(hpa *core.HorizontalPodAutoscaler) string {
	return hpa.GetUID()
}

func (h *horizontalController) enqueueHpa(hpa *core.HorizontalPodAutoscaler) {
	key := h.HpaKeyFunc(hpa)
	h.queue.Enqueue(key)
	logger.HorizontalControllerLogger.Printf("enqueueHpa key %s\n", key)
}

func (h *horizontalController) addHpa(obj interface{}) {
	hpa := obj.(*core.HorizontalPodAutoscaler)
	logger.HorizontalControllerLogger.Printf("Adding %s %s/%s\n", h.Kind, hpa.Namespace, hpa.Name)
	h.enqueueHpa(hpa)
}

func (h *horizontalController) updateHpa(old, cur interface{}) {
	// oldHpa := old.(*core.HorizontalPodAutoscaler)
	curHpa := cur.(*core.HorizontalPodAutoscaler)
	logger.HorizontalControllerLogger.Printf("Updating %s %s/%s\n", h.Kind, curHpa.Namespace, curHpa.Name)

	h.enqueueHpa(curHpa)
}

func (h *horizontalController) deleteHpa(obj interface{}) {
	hpa := obj.(*core.HorizontalPodAutoscaler)
	logger.HorizontalControllerLogger.Printf("Deleting %s, uid %s\n", h.Kind, hpa.UID)

	rssOwned, err := h.getRSOwned(hpa)
	if err != nil {
		logger.HorizontalControllerLogger.Printf("%v\n", err)
	}

	// delete hpa owner ref in owned rss
	h.updatePreOwnedRss(hpa, rssOwned)

	// leave corresponding rs alive instead of killing them
}

const scaleCheckInterval = 15 * time.Second

func (h *horizontalController) periodicallyCheckScale() {
	go h.periodicallyScaleAll()
}

func (h *horizontalController) periodicallyScaleAll() {
	for {
		time.Sleep(scaleCheckInterval)
		hpas := h.hpaInformer.List()

		logger.HorizontalControllerLogger.Printf("[periodicallyScaleAll] enqueue all Hpa start\n")

		for _, item := range hpas {
			hpa := item.(*core.HorizontalPodAutoscaler)
			h.enqueueHpa(hpa)
		}

		logger.HorizontalControllerLogger.Printf("[periodicallyScaleAll] enqueue all Hpa finish\n")
	}
}

func (h *horizontalController) runWorker(ctx context.Context) {
	go h.worker(ctx)
}

const defaultWorkerSleepInterval = time.Duration(3) * time.Second

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (h *horizontalController) worker(ctx context.Context) {

	// go wait.UntilWithContext(ctx, rsc.worker, time.Second)
	for {
		select {
		case <-ctx.Done():
			logger.HorizontalControllerLogger.Printf("[worker] ctx.Done() received, worker of HorizontalController exit\n")
			return
		default:
			for h.processNextWorkItem(ctx) {
			}
			time.Sleep(defaultWorkerSleepInterval)
		}
	}

}

func (h *horizontalController) processNextWorkItem(ctx context.Context) bool {

	item, ok := h.queue.Dequeue()
	if !ok {
		return false
	}

	key := item.(string)

	err := h.reconcileAutoscaler(ctx, key)
	if err != nil {
		logger.HorizontalControllerLogger.Printf("[reconcileAutoscaler] err: %v\n", err)
		// enqueue if error happen when processing
		h.queue.Enqueue(key)
		return false
	}

	return true
}

func (h *horizontalController) reconcileAutoscaler(ctx context.Context, key string) error {

	logger.HorizontalControllerLogger.Printf("[reconcileAutoscaler] start\n")

	// Get the Hpa
	hpaItem, exist := h.hpaInformer.Get(key)
	if !exist {
		return errors.New(fmt.Sprintf("[reconcileAutoscaler] Hpa key: %v is not exist in HpaInformer", key))
	}

	hpa, ok := hpaItem.(*core.HorizontalPodAutoscaler)
	if !ok {
		return errors.New(fmt.Sprintf("[reconcileAutoscaler] key: %v is not Hpa type in HpaInformer", key))
	}

	// Check autoscaling target ScaleTargetRef Kind
	switch hpa.Spec.ScaleTargetRef.Kind {
	case string(types.ReplicasetObjectType):
		// Control replicas num through rs
		rs, exist := h.getScaleTargetRefRS(hpa)
		if !exist {
			return errors.New(fmt.Sprintf("[reconcileAutoscaler] RS name %v not found in rsInformer", hpa.Spec.ScaleTargetRef.Name))
		}

		rescale := false
		rescaleReason := ""
		currentReplicas := rs.Spec.Replicas
		if currentReplicas > hpa.Spec.MaxReplicas {
			rescaleReason = "Current number of replicas above Spec.MaxReplicas"
			rs.Spec.Replicas = hpa.Spec.MaxReplicas
			rescale = true
		} else if currentReplicas < hpa.Spec.MinReplicas {
			rescaleReason = "Current number of replicas below Spec.MinReplicas"
			rs.Spec.Replicas = hpa.Spec.MinReplicas
			rescale = true
		} else if currentReplicas == 0 {
			rescaleReason = "Current number of replicas must be greater than 0"
			rs.Spec.Replicas = 1
			rescale = true
		} else {
			var err error
			rs.Spec.Replicas, rescale, rescaleReason, err = h.calculateDesiredReplicasByMertics(hpa, rs)
			if err != nil {
				return err
			}

			if rs.Spec.Replicas > hpa.Spec.MaxReplicas {
				rs.Spec.Replicas = hpa.Spec.MaxReplicas
			} else if currentReplicas < hpa.Spec.MinReplicas {
				rs.Spec.Replicas = hpa.Spec.MinReplicas
			}
		}

		if rescale && h.rescaleTimeOut(hpa) {

			hpa.Status.LastScaleTime = time.Now()
			hpa.Status.CurrentReplicas = currentReplicas
			hpa.Status.DesiredReplicas = rs.Spec.Replicas

			logger.HorizontalControllerLogger.Printf("[reconcileAutoscaler] rescale start for reason: %v\n", rescaleReason)
			logger.HorizontalControllerLogger.Printf("[reconcileAutoscaler] new status of hpa: CurrentReplicas %v, DesiredReplicas %v\n", hpa.Status.CurrentReplicas, hpa.Status.DesiredReplicas)

			// update hpa
			_, _, err := h.hpaClient.Put(hpa.UID, hpa)
			if err != nil {
				return err
			}

			// update rs
			_, _, err = h.rsClient.Put(rs.UID, rs)
			if err != nil {
				return err
			}

			logger.HorizontalControllerLogger.Printf("[reconcileAutoscaler] rescale finished for reason: %v\n", rescaleReason)
		} else {
			logger.HorizontalControllerLogger.Printf("[reconcileAutoscaler] No rescale this time: %v\n", rescaleReason)
		}

	default:
		return errors.New(fmt.Sprintf("[reconcileAutoscaler] horizontalPodAutoScaler ScaleTargetRef kind %s not support", hpa.Spec.ScaleTargetRef.Kind))
	}
	return nil
}

func (h *horizontalController) getRSOwned(hpa *core.HorizontalPodAutoscaler) (rssOwned []core.ReplicaSet, err error) {
	allRss := h.rsInformer.List()

	hpaUID := hpa.GetUID()

	rssOwned = make([]core.ReplicaSet, 0)
	// calculate actual Replica pod number
	for _, rsItem := range allRss {
		rs, ok := rsItem.(*core.ReplicaSet)
		if !ok {
			return rssOwned, errors.New(fmt.Sprintf("[getRSOwned] Not replicaset type in RsInformer"))
		}

		// check if hpa is rs owner
		if isOwner, owner := meta.CheckOwner(hpaUID, rs.OwnerReferences); isOwner {
			// hpa is owner of this rs
			if meta.CheckOwnerKind(types.HorizontalPodAutoscalerObjectType, owner) {
				rssOwned = append(rssOwned, *rs)
				logger.HorizontalControllerLogger.Printf("[getRSOwned] HPA %v is owner of RS %v\n", hpa.UID, rs.UID)

			} else {
				return rssOwned, errors.New(fmt.Sprintf("[getRSOwned] uid: %v is not HPA type in rs OwnerReferences", hpaUID))
			}
		}
	}

	return rssOwned, nil
}

func (h *horizontalController) calculateDesiredReplicasByMertics(hpa *core.HorizontalPodAutoscaler, rs *core.ReplicaSet) (rsSpecReplicas int32, doRescale bool, rescaleReason string, err error) {

	metricSpecs := hpa.Spec.Metrics
	doRescale = false
	rsSpecReplicas = rs.Spec.Replicas
	err = nil

	currentContainerMetrics, err := h.metricsClient.CollectAllMetrics()
	ownedPods := h.getRSOwnedPods(rs)
	podMetrics := metrics.RearrangeContainerMetricsByPods(currentContainerMetrics, ownedPods)

	cpuRequests := make(map[string]uint64, len(ownedPods))
	memRequests := make(map[string]uint64, len(ownedPods))
	for _, ownedPod := range ownedPods {
		containers := rs.Spec.Template.Spec.Containers
		for _, c := range containers {
			for name, q := range c.Resources.Requests {
				quantity, err := types.ParseQuantity(name, q)
				if err != nil {
					return rs.Spec.Replicas, false, "", err
				}
				switch name {
				case types.ResourceCPU:
					cpuRequests[ownedPod.GetUID()] = quantity
				case types.ResourceMemory:
					memRequests[ownedPod.GetUID()] = quantity
				default:
					continue
				}
			}
		}
	}

	if err != nil {
		logger.HorizontalControllerLogger.Printf("[metricsClient] CollectAllMetrics error: %v\n", err)
		return rs.Spec.Replicas, false, "", errors.New(fmt.Sprintf("[calculateDesiredReplicasByMertics] CollectAllMetrics error: %v", err))
	}

	for _, metric := range metricSpecs {
		switch metric.Type {
		case core.ResourceMetricSourceType:
			if metric.Resource == nil {
				continue
			}
			switch metric.Resource.Target.Type {
			case core.UtilizationMetricType:
				switch metric.Resource.Name {

				case types.ResourceCPU:
					utilizationRatio, currentUtilization, rawAverageValue, err := metrics.GetResourceUtilizationRatio(metric.Resource.Name, podMetrics, cpuRequests, metric.Resource.Target.AverageUtilization)
					logger.HorizontalControllerLogger.Printf("[UtilizationMetricType] GetResourceUtilizationRatio Cpu utilizationRatio %v, currentUtilization: %v, rawAverageValue: %v\n", utilizationRatio, currentUtilization, rawAverageValue)
					if err != nil {
						return rs.Spec.Replicas, false, "", err
					}

					expectedReplicas := int32(math.Ceil(float64(rs.Spec.Replicas) * utilizationRatio))
					if expectedReplicas != rsSpecReplicas {
						rescaleReason = fmt.Sprintf("ResourceCPU Utilization: rsSpecReplicas %v, expectedReplicas %v; utilizationRatio %v, currentUtilization %v", rsSpecReplicas, expectedReplicas, utilizationRatio, currentUtilization)
						rsSpecReplicas = expectedReplicas
						doRescale = true
						return rsSpecReplicas, doRescale, rescaleReason, err
					}

				case types.ResourceMemory:
					utilizationRatio, currentUtilization, rawAverageValue, err := metrics.GetResourceUtilizationRatio(metric.Resource.Name, podMetrics, memRequests, metric.Resource.Target.AverageUtilization)
					logger.HorizontalControllerLogger.Printf("[UtilizationMetricType] GetResourceUtilizationRatio Mem utilizationRatio %v, currentUtilization: %v, rawAverageValue: %v\n", utilizationRatio, currentUtilization, rawAverageValue)
					if err != nil {
						return rs.Spec.Replicas, false, "", err
					}

					expectedReplicas := int32(math.Ceil(float64(rs.Spec.Replicas) * utilizationRatio))
					if expectedReplicas != rsSpecReplicas {
						rescaleReason = fmt.Sprintf("ResourceMem Utilization: rsSpecReplicas %v, expectedReplicas %v; utilizationRatio %v, currentUtilization %v", rsSpecReplicas, expectedReplicas, utilizationRatio, currentUtilization)
						rsSpecReplicas = expectedReplicas
						doRescale = true
						return rsSpecReplicas, doRescale, rescaleReason, err
					}

				default:
					logger.HorizontalControllerLogger.Printf("[calculateDesiredReplicasByMertics] resource name type %v not supported\n", metric.Resource.Name)
				}
			case core.AverageValueMetricType:
				switch metric.Resource.Name {
				case types.ResourceCPU, types.ResourceMemory:
					avgVal, err := types.ParseQuantity(metric.Resource.Name, metric.Resource.Target.AverageValue)
					if err != nil {
						return rs.Spec.Replicas, false, "", err
					}

					usageRatio, currentUsage := metrics.GetMetricUsageRatio(metric.Resource.Name, podMetrics, avgVal)
					logger.HorizontalControllerLogger.Printf("[UtilizationMetricType] GetMetricUsageRatio %v, usageRatio %v, currentUsage: %v\n", metric.Resource.Name, usageRatio, currentUsage)

					expectedReplicas := int32(math.Ceil(float64(rs.Spec.Replicas) * usageRatio))
					if expectedReplicas != rsSpecReplicas {
						rescaleReason = fmt.Sprintf("Resource %v AverageValue: rsSpecReplicas %v, expectedReplicas %v; usageRatio %v, currentUsage %v", metric.Resource.Name, rsSpecReplicas, expectedReplicas, usageRatio, currentUsage)
						rsSpecReplicas = expectedReplicas
						doRescale = true
						return rsSpecReplicas, doRescale, rescaleReason, err
					}
				default:
					logger.HorizontalControllerLogger.Printf("[calculateDesiredReplicasByMertics] resource name type %v not supported\n", metric.Resource.Name)
				}
			default:
				logger.HorizontalControllerLogger.Printf("[calculateDesiredReplicasByMertics] resource target type %v not supported\n", metric.Resource.Target.Type)
			}
		default:
			return rs.Spec.Replicas, false, "", errors.New(fmt.Sprintf("[calculateDesiredReplicasByMertics] mertic %v not supported", metric.Type))
		}
	}

	return rsSpecReplicas, doRescale, rescaleReason, err
}

const defaultRescaleTimeInterval = time.Duration(5) * time.Second

func (h *horizontalController) rescaleTimeOut(hpa *core.HorizontalPodAutoscaler) bool {
	return time.Since(hpa.Status.LastScaleTime) > defaultRescaleTimeInterval
}

func (h *horizontalController) updatePreOwnedRss(hpa *core.HorizontalPodAutoscaler, preOwned []core.ReplicaSet) {
	for _, rs := range preOwned {
		// Delete hpa owner reference
		rs.DeleteOwnerReference(hpa.UID)

		// Ask ApiServer to update rs
		_, _, err := h.rsClient.Put(rs.UID, &rs)
		if err != nil {
			logger.HorizontalControllerLogger.Printf("[updatePreOwnedRss] Put failed when ask ApiServer to update rs uid %v, err: %v\n", rs.UID, err)
		}
	}
}

func (h *horizontalController) getScaleTargetRefRS(hpa *core.HorizontalPodAutoscaler) (*core.ReplicaSet, bool) {
	scaleTargetRef := hpa.Spec.ScaleTargetRef
	allRss := h.rsInformer.List()

	for _, rsItem := range allRss {
		rs, ok := rsItem.(*core.ReplicaSet)
		if !ok {
			return nil, false
		}

		if rs.Name == scaleTargetRef.Name && rs.APIVersion == scaleTargetRef.APIVersion {
			return rs, true
		}
	}

	return nil, false
}

func (h *horizontalController) getRSOwnedPods(rs *core.ReplicaSet) []core.Pod {
	rsUID := rs.GetUID()
	relatedPods := make([]core.Pod, 0)
	pods := h.podInformer.List()
	for _, item := range pods {
		pod := item.(*core.Pod)
		if isOwner, owner := meta.CheckOwner(rsUID, pod.OwnerReferences); isOwner && meta.CheckOwnerKind(types.ReplicasetObjectType, owner) {
			relatedPods = append(relatedPods, *pod)
		}
	}

	return relatedPods
}

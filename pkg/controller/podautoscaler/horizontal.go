package podautoscaler

import (
	"context"
	"errors"
	"fmt"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/types"
	client "minik8s/pkg/apiclient/interface"
	"minik8s/pkg/controller/cache"
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
		podClient:   podClient,
		hpaClient:   hpaClient,
		rsClient:    rsClient,
		podInformer: podInformer,
		hpaInformer: hpaInformer,
		rsInformer:  rsInformer,
		queue:       cache.NewWorkQueue(),
		Kind:        string(types.HorizontalPodAutoscalerObjectType),
	}

	_ = hc.hpaInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    hc.addHpa,
		UpdateFunc: hc.updateHpa,
		DeleteFunc: hc.deleteHpa,
	})

	return hc
}

type horizontalController struct {
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

	logger.HorizontalControllerLogger.Printf("[HorizontalController] start\n")

	h.runWorker(ctx)

	// wait for controller manager stop
	<-ctx.Done()
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
		}

		if rescale && h.rescaleTimeOut(hpa) {

			hpa.Status.LastScaleTime = time.Now()
			hpa.Status.CurrentReplicas = currentReplicas
			hpa.Status.DesiredReplicas = rs.Spec.Replicas

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
	metrics := hpa.Spec.Metrics
	doRescale = false
	rsSpecReplicas = rs.Spec.Replicas
	err = nil
	for _, metric := range metrics {
		switch metric.Type {
		case core.ResourceMetricSourceType:
			if metric.Resource == nil {
				continue
			}
			switch metric.Resource.Target.Type {
			case core.UtilizationMetricType:
				switch metric.Resource.Name {
				case types.ResourceCPU:

				case types.ResourceMemory:

				default:
					logger.HorizontalControllerLogger.Printf("[calculateDesiredReplicasByMertics] resource name type %v not supported\n", metric.Resource.Name)
				}
			case core.AverageValueMetricType:
				switch metric.Resource.Name {
				case types.ResourceCPU:

				case types.ResourceMemory:

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

func (h *horizontalController) rescaleTimeOut(hpa *core.HorizontalPodAutoscaler) bool {
	// TODO
	return true
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

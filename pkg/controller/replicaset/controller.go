package replicaset

import (
	"context"
	"errors"
	"fmt"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/generate"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/types"
	client "minik8s/pkg/apiclient/interface"
	"minik8s/pkg/controller/cache"
	"minik8s/pkg/logger"
	"reflect"
	"time"
)

type ReplicaSetController interface {
	Run(ctx context.Context)
}

func NewReplicaSetController(podInformer cache.Informer, podClient client.Interface, rsInformer cache.Informer, rsClient client.Interface) ReplicaSetController {

	rsc := &replicaSetController{
		Kind:        string(types.ReplicasetObjectType),
		PodInformer: podInformer,
		PodClient:   podClient,
		RsInformer:  rsInformer,
		RsClient:    rsClient,
		queue:       cache.NewWorkQueue(),
	}

	_ = rsc.RsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    rsc.addRS,
		UpdateFunc: rsc.updateRS,
		DeleteFunc: rsc.deleteRS,
	})

	_ = rsc.PodInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    rsc.addPod,
		UpdateFunc: rsc.updatePod,
		DeleteFunc: rsc.deletePod,
	})

	return rsc
}

type replicaSetController struct {
	Kind string

	PodInformer cache.Informer
	RsInformer  cache.Informer
	PodClient   client.Interface
	RsClient    client.Interface
	queue       cache.WorkQueue
}

func (rsc *replicaSetController) Run(ctx context.Context) {

	go func() {
		logger.ReplicaSetControllerLogger.Printf("[ReplicaSetController] start\n")
		defer logger.ReplicaSetControllerLogger.Printf("[ReplicaSetController] finish\n")

		rsc.runWorker(ctx)

		// wait for controller manager stop
		<-ctx.Done()
	}()
	return
}

func (rsc *replicaSetController) RSKeyFunc(rs *core.ReplicaSet) string {
	return rs.GetUID()
}

func (rsc *replicaSetController) enqueueRS(rs *core.ReplicaSet) {
	key := rsc.RSKeyFunc(rs)
	rsc.queue.Enqueue(key)
	logger.ReplicaSetControllerLogger.Printf("enqueueRS key %s\n", key)
}

func (rsc *replicaSetController) addRS(obj interface{}) {
	rs := obj.(*core.ReplicaSet)
	logger.ReplicaSetControllerLogger.Printf("Adding %s %s/%s\n", rsc.Kind, rs.Namespace, rs.Name)
	rsc.enqueueRS(rs)
}

// callback when RS is updated
func (rsc *replicaSetController) updateRS(old, cur interface{}) {
	oldRS := old.(*core.ReplicaSet)
	curRS := cur.(*core.ReplicaSet)

	//if curRS.UID != oldRS.UID {
	//	key := rsc.RSKeyFunc(oldRS)
	//
	//	rsc.deleteRS(cache.DeletedFinalStateUnknown{
	//		Key: key,
	//		Obj: oldRS,
	//	})
	//}

	if (oldRS.Spec.Replicas) != (curRS.Spec.Replicas) {
		logger.ReplicaSetControllerLogger.Printf("%v %v updated uid %v. Desired pod count change: %d->%d\n", rsc.Kind, curRS.Name, curRS.UID, oldRS.Spec.Replicas, curRS.Spec.Replicas)
	}
	rsc.enqueueRS(curRS)
}

func (rsc *replicaSetController) deleteRS(obj interface{}) {
	rs := obj.(*core.ReplicaSet)

	logger.ReplicaSetControllerLogger.Printf("Deleting %s, uid %s\n", rsc.Kind, rs.UID)

	podsOwned, podsPreOwned, err := rsc.getPodsOwned(rs)
	if err != nil {
		logger.ReplicaSetControllerLogger.Printf("%v\n", err)
	}

	// delete rs owner ref in preowned pods
	rsc.updatePreOwnedPods(rs, podsPreOwned)

	// Delete pods of according rs
	// TODO: check owner reference of pod in case it has other owner, meaning pod can not be delete
	err = rsc.decreasePods(int32(len(podsOwned)), podsOwned)
	if err != nil {
		logger.ReplicaSetControllerLogger.Printf("[deleteRS] Delete failed when ask ApiServer to delete pods, %v\n", err)
		return
	}

}

// When a pod is created, enqueue the replica set that manages it
func (rsc *replicaSetController) addPod(obj interface{}) {
	pod := obj.(*core.Pod)

	rss := rsc.selectReplicaSetMatchesLabel(pod.Labels)
	if len(rss) == 0 {
		return
	}
	for _, rs := range rss {
		logger.ReplicaSetControllerLogger.Printf("enqueue ReplicaSet: uid %s when add Pod happen: uid %s\n", rs.UID, pod.UID)
		rsc.enqueueRS(rs)
	}
}

// When a pod is updated, figure out what replica set/s manage it and wake them
// up. If the labels of the pod have changed we need to awaken both the old
// and new replica set. old and cur must be *v1.Pod types.
func (rsc *replicaSetController) updatePod(old, cur interface{}) {
	curPod := cur.(*core.Pod)
	oldPod := old.(*core.Pod)

	if curPod.ResourceVersion == oldPod.ResourceVersion {
		// Periodic resync will send update events for all known pods.
		// Two different versions of the same pod will always have different RVs.
		return
	}

	labelChanged := !reflect.DeepEqual(curPod.Labels, oldPod.Labels)
	if labelChanged {

		// check if old pod owner has rs
		rs := rsc.getPodOwnerReplicaSet(oldPod)
		if rs != nil {
			logger.ReplicaSetControllerLogger.Printf("enqueue ReplicaSet %s when update Pod %s\n", rs.UID, curPod.UID)
			rsc.enqueueRS(rs)
		}

		logger.ReplicaSetControllerLogger.Printf("label changed for update Pod %s\n", curPod.UID)
		rss := rsc.selectReplicaSetMatchesLabel(curPod.Labels)
		if len(rss) == 0 {
			return
		}
		for _, rs = range rss {
			logger.ReplicaSetControllerLogger.Printf("enqueue ReplicaSet %s when update Pod %s\n", rs.UID, curPod.UID)
			rsc.enqueueRS(rs)
		}
	}

}

// When a pod is deleted, enqueue the replica set that manages the pod
func (rsc *replicaSetController) deletePod(obj interface{}) {
	pod := obj.(*core.Pod)

	rs := rsc.getPodOwnerReplicaSet(pod)
	if rs != nil {
		logger.ReplicaSetControllerLogger.Printf("enqueue ReplicaSet %s when delete Pod %s\n", rs.UID, pod.UID)
		rsc.enqueueRS(rs)
	} else {
		logger.ReplicaSetControllerLogger.Printf("delete Pod %s, which has no owner ReplicaSet\n", pod.UID)
	}
}

func (rsc *replicaSetController) runWorker(ctx context.Context) {
	go rsc.worker(ctx)
}

const defaultWorkerSleepInterval = time.Duration(3) * time.Second

// worker runs a worker thread that just dequeues items, processes them, and marks them done.
// It enforces that the syncHandler is never invoked concurrently with the same key.
func (rsc *replicaSetController) worker(ctx context.Context) {

	// go wait.UntilWithContext(ctx, rsc.worker, time.Second)
	for {
		select {
		case <-ctx.Done():
			logger.ReplicaSetControllerLogger.Printf("[worker] ctx.Done() received, worker of ReplicaSetController exit\n")
			return
		default:
			for rsc.processNextWorkItem(ctx) {
			}
			time.Sleep(defaultWorkerSleepInterval)
		}
	}

}

func (rsc *replicaSetController) processNextWorkItem(ctx context.Context) bool {

	item, ok := rsc.queue.Dequeue()
	if !ok {
		return false
	}

	key := item.(string)

	err := rsc.syncReplicaSet(ctx, key)
	if err != nil {
		logger.ReplicaSetControllerLogger.Printf("[syncReplicaSet] err: %v\n", err)
		// enqueue if error happen when processing
		rsc.queue.Enqueue(key)
		return false
	}

	return true
}

// syncReplicaSet will sync the ReplicaSet with the given key if it has had its expectations fulfilled,
// meaning it did not expect to see any more of its pods created or deleted. This function is not meant to be
// invoked concurrently with the same key.
// param key is the uid of object
func (rsc *replicaSetController) syncReplicaSet(ctx context.Context, key string) error {

	// Get the ReplicaSet
	rsItem, exist := rsc.RsInformer.Get(key)
	if !exist {
		return errors.New(fmt.Sprintf("[syncReplicaSet] ReplicaSet key: %v is not exist in RsInformer", key))
	}

	rs, ok := rsItem.(*core.ReplicaSet)
	if !ok {
		return errors.New(fmt.Sprintf("[syncReplicaSet] key: %v is not ReplicaSet type in RsInformer", key))
	}

	podsOwned, matchedNotOwnedPods, podsPreOwned, err := rsc.getPodsOwnedAndMatchedNotOwned(rs, true)
	if err != nil {
		logger.ReplicaSetControllerLogger.Printf("[syncReplicaSet] %v\n", err)
		return err
	}

	// delete rs owner ref in preowned pods
	rsc.updatePreOwnedPods(rs, podsPreOwned)

	// count actual replica num of rs
	actualReplicaNum := int32(len(podsOwned))
	logger.ReplicaSetControllerLogger.Printf("[syncReplicaSet] Actual replica num of rs: %v, expected: %v\n", actualReplicaNum, rs.Spec.Replicas)

	// update ReplicaSet status
	rs.Status.Replicas = actualReplicaNum

	// check if status and spec match
	if rs.Status.Replicas != rs.Spec.Replicas {
		if rs.Status.Replicas < rs.Spec.Replicas {
			rsc.increaseReplica(rs, matchedNotOwnedPods)
		} else {
			rsc.decreaseReplica(rs, podsOwned)
		}
	}

	return nil
}

func (rsc *replicaSetController) getPodsOwned(rs *core.ReplicaSet) (podsOwned []core.Pod, podsPreOwned []core.Pod, err error) {
	podsOwned, _, podsPreOwned, err = rsc.getPodsOwnedAndMatchedNotOwned(rs, false)
	return podsOwned, podsPreOwned, err
}

func (rsc *replicaSetController) getPodsOwnedAndMatchedNotOwned(rs *core.ReplicaSet, getMatchedNotOwned bool) (podsOwned []core.Pod, matchedNotOwnedPods []core.Pod, podsPreOwned []core.Pod, err error) {
	allPods := rsc.PodInformer.List()

	rsUID := rs.GetUID()

	podsOwned = make([]core.Pod, 0)
	matchedNotOwnedPods = make([]core.Pod, 0)
	podsPreOwned = make([]core.Pod, 0)
	// calculate actual Replica pod number
	for _, podItem := range allPods {

		pod, ok := podItem.(*core.Pod)
		if !ok {
			return podsOwned, matchedNotOwnedPods, podsPreOwned, errors.New(fmt.Sprintf("[getPodsOwnedAndMatchedNotOwned] Not Pod type in PodInformer"))
		}

		// check if rs is pod owner
		if isOwner, owner := meta.CheckOwner(rsUID, pod.OwnerReferences); isOwner {

			// rs is owner of this pod
			if meta.CheckOwnerKind(types.ReplicasetObjectType, owner) {
				if meta.MatchLabelSelector(rs.Spec.Selector, pod.Labels) {
					// label still match
					// append pod info to podsOwned
					podsOwned = append(podsOwned, *pod)
					logger.ReplicaSetControllerLogger.Printf("[getPodsOwnedAndMatchedNotOwned] rs %v is owner of pod %v\n", rs.UID, pod.UID)

				} else {
					// label change and not match
					// append pod info to podsPreOwned
					podsPreOwned = append(podsPreOwned, *pod)
					logger.ReplicaSetControllerLogger.Printf("[getPodsOwnedAndMatchedNotOwned] rs %v is previous the owner of pod %v, but not label changed\n", rs.UID, pod.UID)
				}

			} else {
				return podsOwned, matchedNotOwnedPods, podsPreOwned, errors.New(fmt.Sprintf("[getPodsOwnedAndMatchedNotOwned] uid: %v is not ReplicaSet type in pod OwnerReferences", rsUID))
			}

		} else {
			// rs is not owner of this pod
			hasRsOwner, _ := meta.HasOwnerKind(types.ReplicasetObjectType, pod.OwnerReferences)
			if getMatchedNotOwned && !hasRsOwner && meta.MatchLabelSelector(rs.Spec.Selector, pod.Labels) {
				// label selector match and pod don't have rs owner
				matchedNotOwnedPods = append(matchedNotOwnedPods, *pod)
				logger.ReplicaSetControllerLogger.Printf("[getPodsOwnedAndMatchedNotOwned] label selector match and pod %v don't have rs owner\n", pod.UID)
			}
		}
	}

	return podsOwned, matchedNotOwnedPods, podsPreOwned, nil
}

func (rsc *replicaSetController) increaseReplica(rs *core.ReplicaSet, matchedNotOwnedPods []core.Pod) {

	var podReplicaNum = rs.Status.Replicas

	numToIncrease := rs.Spec.Replicas - rs.Status.Replicas
	modifyNotOwnedPodsNum := int32(len(matchedNotOwnedPods))
	if numToIncrease < modifyNotOwnedPodsNum {
		modifyNotOwnedPodsNum = numToIncrease
	}

	// Generate rs ownerReference
	ownerReference := rs.GenerateOwnerReference()

	var idx int32
	// add pods matched labels
	for idx = 0; idx < modifyNotOwnedPodsNum; idx++ {
		pod := matchedNotOwnedPods[idx]
		pod.AppendOwnerReference(ownerReference)

		// Ask ApiServer to update pod
		_, _, err := rsc.PodClient.Put(pod.UID, &pod)
		if err != nil {
			logger.ReplicaSetControllerLogger.Printf("[increaseReplica] Put failed when ask ApiServer to update pod uid %v, err: %v\n", pod.UID, err)
			rsc.finishModifyReplicaAndUpdateRsStatus(rs, podReplicaNum)
			return
		}

		podReplicaNum++
	}

	// create new pods from template
	for idx = 0; idx < numToIncrease-int32(len(matchedNotOwnedPods)); idx++ {

		// Generate new pod from rs pod template
		newPod := generate.PodFromReplicaSet(rs)
		newPod.AppendOwnerReference(ownerReference)

		// Ask ApiServer to create new pods
		_, postResponse, err := rsc.PodClient.Post(newPod)
		if err != nil {
			logger.ReplicaSetControllerLogger.Printf("[increaseReplica] Post failed when ask ApiServer to create new pods, %v\n", err)
			rsc.finishModifyReplicaAndUpdateRsStatus(rs, podReplicaNum)
			return
		}

		// Wait for create of pod finish
		logger.ReplicaSetControllerLogger.Printf("[increaseReplica] New Pod name %s successfully created, uid %v\n", newPod.Name, postResponse.UID)
		podReplicaNum++
	}

	rsc.finishModifyReplicaAndUpdateRsStatus(rs, podReplicaNum)

}

func (rsc *replicaSetController) finishModifyReplicaAndUpdateRsStatus(rs *core.ReplicaSet, replicaNum int32) {
	rs.Status.Replicas = replicaNum
	if rs.Status.Replicas != rs.Spec.Replicas {
		logger.ReplicaSetControllerLogger.Printf("[finishModifyReplicaAndUpdateRsStatus] Something wrong, rs.Status.Replicas != rs.Spec.Replicas after increase/decrease Replica finished!\n")
		return
	}

	// update rs status
	_, _, err := rsc.RsClient.Put(rs.UID, rs)
	if err != nil {
		logger.ReplicaSetControllerLogger.Printf("[finishModifyReplicaAndUpdateRsStatus] Put failed when ask ApiServer to update rs, %v\n", err)
		return
	}
}

func (rsc *replicaSetController) decreaseReplica(rs *core.ReplicaSet, podsOwned []core.Pod) {
	numToDecrease := rs.Status.Replicas - rs.Spec.Replicas

	// ask ApiServer to delete pods
	err := rsc.decreasePods(numToDecrease, podsOwned)
	if err != nil {
		logger.ReplicaSetControllerLogger.Printf("[decreaseReplica] Delete failed when ask ApiServer to delete new pods, %v\n", err)
		return
	}

	rsc.finishModifyReplicaAndUpdateRsStatus(rs, rs.Spec.Replicas)
}

// decreasePods ask ApiServer to delete numToDecrease pods in podsOwned
func (rsc *replicaSetController) decreasePods(numToDecrease int32, podsOwned []core.Pod) error {
	// TODO: check owner reference of pod in case it has other owner, meaning pod can not be delete

	// ask ApiServer to delete pods
	var idx int32
	for idx = 0; idx < numToDecrease; idx++ {
		podToDelete := podsOwned[idx]

		// delete pod
		_, _, err := rsc.PodClient.Delete(podToDelete.UID)
		if err != nil {
			logger.ReplicaSetControllerLogger.Printf("[decreasePods] Delete failed when ask ApiServer to delete pod %v, %v\n", podToDelete.UID, err)
			return err
		}
	}

	return nil
}

// selectReplicaSetMatchesLabel return ReplicaSets that matches the labels (of pod)
func (rsc *replicaSetController) selectReplicaSetMatchesLabel(labels map[string]string) []*core.ReplicaSet {
	rss := rsc.RsInformer.List()
	var result []*core.ReplicaSet

	for _, item := range rss {
		rs := item.(*core.ReplicaSet)

		matches := meta.MatchLabelSelector(rs.Spec.Selector, labels)

		if matches {
			result = append(result, rs)
		}
	}

	return result
}

func (rsc *replicaSetController) getPodOwnerReplicaSet(pod *core.Pod) (rs *core.ReplicaSet) {
	hasRsOwner, owner := meta.HasOwnerKind(types.ReplicasetObjectType, pod.OwnerReferences)
	if !hasRsOwner {
		return nil
	} else {
		rsItem, exist := rsc.RsInformer.Get(owner.UID)
		if exist {
			rs = rsItem.(*core.ReplicaSet)
			return rs
		} else {
			return nil
		}
	}
}

func (rsc *replicaSetController) updatePreOwnedPods(rs *core.ReplicaSet, preOwned []core.Pod) {

	for _, p := range preOwned {

		// Delete rs owner reference
		p.DeleteOwnerReference(rs.UID)

		// Ask ApiServer to update pod
		_, _, err := rsc.PodClient.Put(p.UID, &p)
		if err != nil {
			logger.ReplicaSetControllerLogger.Printf("[updatePreOwnedPods] Put failed when ask ApiServer to update pod uid %v, err: %v\n", p.UID, err)
		}
	}
}

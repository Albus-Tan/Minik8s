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
	"time"
)

type ReplicaSetController interface {
	Run(ctx context.Context)
}

func NewReplicaSetController(podInformer cache.Informer, podClient client.Interface, rsInformer cache.Informer, rsClient client.Interface) ReplicaSetController {

	rsc := &replicaSetController{
		Kind:        "ReplicaSet",
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
	//TODO implement me

	// logger.ReplicaSetControllerLogger.SetPrefix()
	logger.ReplicaSetControllerLogger.Printf("[ReplicaSetController] start\n")

	rsc.runWorker(ctx)

	// wait for controller manager stop
	<-ctx.Done()

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

	// Delete expectations for the ReplicaSet so if we create a new one with the same name it starts clean

}

// When a pod is created, enqueue the replica set that manages it and update its expectations.
func (rsc *replicaSetController) addPod(obj interface{}) {
	//pod := obj.(*core.Pod)
	//
	//// If it has a ControllerRef, that's all that matters.
	//if controllerRef := metav1.GetControllerOf(pod); controllerRef != nil {
	//	rs := rsc.resolveControllerRef(pod.Namespace, controllerRef)
	//	if rs == nil {
	//		return
	//	}
	//	rsKey := rsc.RSKeyFunc(rs)
	//	if err != nil {
	//		return
	//	}
	//	klogger.ReplicaSetControllerLogger.V(4).Infof("Pod %s created: %#v.", pod.Name, pod)
	//	rsc.expectations.CreationObserved(rsKey)
	//	rsc.queue.Add(rsKey)
	//	return
	//}
	//
	//// Otherwise, it's an orphan. Get a list of all matching ReplicaSets and sync
	//// them to see if anyone wants to adopt it.
	//// DO NOT observe creation because no controller should be waiting for an
	//// orphan.
	//rss := rsc.getPodReplicaSets(pod)
	//if len(rss) == 0 {
	//	return
	//}
	//klogger.ReplicaSetControllerLogger.V(4).Infof("Orphan Pod %s created: %#v.", pod.Name, pod)
	//for _, rs := range rss {
	//	rsc.enqueueRS(rs)
	//}
}

// When a pod is updated, figure out what replica set/s manage it and wake them
// up. If the labels of the pod have changed we need to awaken both the old
// and new replica set. old and cur must be *v1.Pod types.
func (rsc *replicaSetController) updatePod(old, cur interface{}) {
	//curPod := cur.(*v1.Pod)
	//oldPod := old.(*v1.Pod)
	//if curPod.ResourceVersion == oldPod.ResourceVersion {
	//	// Periodic resync will send update events for all known pods.
	//	// Two different versions of the same pod will always have different RVs.
	//	return
	//}
	//
	//labelChanged := !reflect.DeepEqual(curPod.Labels, oldPod.Labels)
	//if curPod.DeletionTimestamp != nil {
	//	// when a pod is deleted gracefully it's deletion timestamp is first modified to reflect a grace period,
	//	// and after such time has passed, the kubelet actually deletes it from the store. We receive an update
	//	// for modification of the deletion timestamp and expect an rs to create more replicas asap, not wait
	//	// until the kubelet actually deletes the pod. This is different from the Phase of a pod changing, because
	//	// an rs never initiates a phase change, and so is never asleep waiting for the same.
	//	rsc.deletePod(curPod)
	//	if labelChanged {
	//		// we don't need to check the oldPod.DeletionTimestamp because DeletionTimestamp cannot be unset.
	//		rsc.deletePod(oldPod)
	//	}
	//	return
	//}
	//
	//curControllerRef := metav1.GetControllerOf(curPod)
	//oldControllerRef := metav1.GetControllerOf(oldPod)
	//controllerRefChanged := !reflect.DeepEqual(curControllerRef, oldControllerRef)
	//if controllerRefChanged && oldControllerRef != nil {
	//	// The ControllerRef was changed. Sync the old controller, if any.
	//	if rs := rsc.resolveControllerRef(oldPod.Namespace, oldControllerRef); rs != nil {
	//		rsc.enqueueRS(rs)
	//	}
	//}
	//
	//// If it has a ControllerRef, that's all that matters.
	//if curControllerRef != nil {
	//	rs := rsc.resolveControllerRef(curPod.Namespace, curControllerRef)
	//	if rs == nil {
	//		return
	//	}
	//	klogger.ReplicaSetControllerLogger.V(4).Infof("Pod %s updated, objectMeta %+v -> %+v.", curPod.Name, oldPod.ObjectMeta, curPod.ObjectMeta)
	//	rsc.enqueueRS(rs)
	//	// TODO: MinReadySeconds in the Pod will generate an Available condition to be added in
	//	// the Pod status which in turn will trigger a requeue of the owning replica set thus
	//	// having its status updated with the newly available replica. For now, we can fake the
	//	// update by resyncing the controller MinReadySeconds after the it is requeued because
	//	// a Pod transitioned to Ready.
	//	// Note that this still suffers from #29229, we are just moving the problem one level
	//	// "closer" to kubelet (from the deployment to the replica set controller).
	//	if !podutil.IsPodReady(oldPod) && podutil.IsPodReady(curPod) && rs.Spec.MinReadySeconds > 0 {
	//		klogger.ReplicaSetControllerLogger.V(2).Infof("%v %q will be enqueued after %ds for availability check", rsc.Kind, rs.Name, rs.Spec.MinReadySeconds)
	//		// Add a second to avoid milliseconds skew in AddAfter.
	//		// See https://github.com/kubernetes/kubernetes/issues/39785#issuecomment-279959133 for more info.
	//		rsc.enqueueRSAfter(rs, (time.Duration(rs.Spec.MinReadySeconds)*time.Second)+time.Second)
	//	}
	//	return
	//}
	//
	//// Otherwise, it's an orphan. If anything changed, sync matching controllers
	//// to see if anyone wants to adopt it now.
	//if labelChanged || controllerRefChanged {
	//	rss := rsc.getPodReplicaSets(curPod)
	//	if len(rss) == 0 {
	//		return
	//	}
	//	klogger.ReplicaSetControllerLogger.V(4).Infof("Orphan Pod %s updated, objectMeta %+v -> %+v.", curPod.Name, oldPod.ObjectMeta, curPod.ObjectMeta)
	//	for _, rs := range rss {
	//		rsc.enqueueRS(rs)
	//	}
	//}
}

// When a pod is deleted, enqueue the replica set that manages the pod and update its expectations.
// obj could be an *v1.Pod, or a DeletionFinalStateUnknown marker item.
func (rsc *replicaSetController) deletePod(obj interface{}) {
	//pod, ok := obj.(*v1.Pod)
	//
	//// When a delete is dropped, the relist will notice a pod in the store not
	//// in the list, leading to the insertion of a tombstone object which contains
	//// the deleted key/value. Note that this value might be stale. If the pod
	//// changed labels the new ReplicaSet will not be woken up till the periodic resync.
	//if !ok {
	//	tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
	//	if !ok {
	//		utilruntime.HandleError(fmt.Errorf("couldn't get object from tombstone %+v", obj))
	//		return
	//	}
	//	pod, ok = tombstone.Obj.(*v1.Pod)
	//	if !ok {
	//		utilruntime.HandleError(fmt.Errorf("tombstone contained object that is not a pod %#v", obj))
	//		return
	//	}
	//}
	//
	//controllerRef := metav1.GetControllerOf(pod)
	//if controllerRef == nil {
	//	// No controller should care about orphans being deleted.
	//	return
	//}
	//rs := rsc.resolveControllerRef(pod.Namespace, controllerRef)
	//if rs == nil {
	//	return
	//}
	//rsKey, err := controller.KeyFunc(rs)
	//if err != nil {
	//	utilruntime.HandleError(fmt.Errorf("couldn't get key for object %#v: %v", rs, err))
	//	return
	//}
	//klogger.ReplicaSetControllerLogger.V(4).Infof("Pod %s/%s deleted through %v, timestamp %+v: %#v.", pod.Namespace, pod.Name, utilruntime.GetCaller(), pod.DeletionTimestamp, pod)
	//rsc.expectations.DeletionObserved(rsKey, controller.PodKey(pod))
	//rsc.queue.Add(rsKey)
}

func (rsc *replicaSetController) runWorker(ctx context.Context) {
	go rsc.worker(ctx)
}

const defaultWorkerSleepInterval = time.Duration(5) * time.Second

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
			rsc.processNextWorkItem(ctx)
			time.Sleep(defaultWorkerSleepInterval)
		}
	}

}

func (rsc *replicaSetController) processNextWorkItem(ctx context.Context) bool {
	// TODO

	//key, quit := rsc.queue.Get()
	//if quit {
	//	return false
	//}
	//defer rsc.queue.Done(key)
	//
	//err := rsc.syncHandler(ctx, key.(string))
	//if err == nil {
	//	rsc.queue.Forget(key)
	//	return true
	//}
	//
	//utilruntime.HandleError(fmt.Errorf("sync %q failed with %v", key, err))
	//rsc.queue.AddRateLimited(key)
	//
	//return true
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

	allPods := rsc.PodInformer.List()

	var actualReplicaNum int32 = 0
	rsUID := rs.GetUID()

	podsOwned := make([]core.Pod, 0)
	// calculate actual Replica pod number
	for _, podItem := range allPods {

		pod, ok := podItem.(*core.Pod)
		if !ok {
			return errors.New(fmt.Sprintf("[syncReplicaSet] Not Pod type in PodInformer"))
		}

		// check if pod have owner rs
		if isOwner, owner := meta.CheckOwner(rsUID, pod.OwnerReferences); isOwner {

			// rs is owner of this pod
			if meta.CheckOwnerKind(types.ReplicasetObjectType, owner) {
				// increase actual replica num of rs
				actualReplicaNum++
				// append pod info to podsOwned
				podsOwned = append(podsOwned, *pod)
			} else {
				return errors.New(fmt.Sprintf("[syncReplicaSet] key: %v is not ReplicaSet type in pod OwnerReferences", key))
			}

		} else {
			// rs is not owner of this pod
			// TODO: match label selector
		}
	}

	// update ReplicaSet status
	rs.Status.Replicas = actualReplicaNum

	// check if status and spec match
	if rs.Status.Replicas != rs.Spec.Replicas {
		if rs.Status.Replicas < rs.Spec.Replicas {
			rsc.increaseReplica(rs)
		} else {
			rsc.decreaseReplica(rs, podsOwned)
		}
	}

	return nil
}

func (rsc *replicaSetController) increaseReplica(rs *core.ReplicaSet) {
	numToIncrease := rs.Spec.Replicas - rs.Status.Replicas

	// Generate rs ownerReference
	ownerReference := rs.GenerateOwnerReference()

	var idx int32
	for idx = 0; idx < numToIncrease; idx++ {

		// Generate new pod from rs pod template
		newPod := generate.PodFromReplicaSet(rs)
		newPod.AppendOwnerReference(ownerReference)

		// Ask ApiServer to create new pods
		_, postResponse, err := rsc.PodClient.Post(newPod)
		if err != nil {
			logger.ReplicaSetControllerLogger.Printf("[increaseReplica] Post failed when ask ApiServer to create new pods, %v\n", err)
			return
		}

		// Wait for create of pod finish
		logger.ReplicaSetControllerLogger.Printf("[increaseReplica] New Pod name %s successfully created, uid %v\n", newPod.Name, postResponse.UID)

	}

	rs.Status.Replicas = rs.Spec.Replicas

	// update rs status
	_, _, err := rsc.RsClient.Put(rs.UID, rs)
	if err != nil {
		logger.ReplicaSetControllerLogger.Printf("[increaseReplica] Put failed when ask ApiServer to update rs, %v\n", err)
		return
	}

}

func (rsc *replicaSetController) decreaseReplica(rs *core.ReplicaSet, podsOwned []core.Pod) {
	numToDecrease := rs.Status.Replicas - rs.Spec.Replicas

	// ask ApiServer to delete pods
	var idx int32
	for idx = 0; idx < numToDecrease; idx++ {
		podToDelete := podsOwned[idx]

		// delete pod
		_, _, err := rsc.PodClient.Delete(podToDelete.UID)
		if err != nil {
			logger.ReplicaSetControllerLogger.Printf("[decreaseReplica] Delete failed when ask ApiServer to delete new pods, %v\n", err)
			return
		}
	}

	rs.Status.Replicas = rs.Spec.Replicas

	// update rs status
	_, _, err := rsc.RsClient.Put(rs.UID, rs)
	if err != nil {
		logger.ReplicaSetControllerLogger.Printf("[decreaseReplica] Put failed when ask ApiServer to update rs, %v\n", err)
		return
	}
}

package serverless

import (
	"context"
	"math"
	"minik8s/config"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/generate"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/types"
	client "minik8s/pkg/apiclient/interface"
	"minik8s/pkg/controller/cache"
	"minik8s/pkg/logger"
	"minik8s/utils"
	"reflect"
	"time"
)

type ServerlessController interface {
	Run(ctx context.Context)
}

func NewServerlessController(
	FuncTemplateInformer cache.Informer,
	FuncTemplateClient client.Interface,
	ReplicaSetClient client.Interface,
	ServiceClient client.Interface,
	PodClient client.Interface,
) ServerlessController {

	sc := &serverlessController{
		Kind:                 string(types.FuncTemplateObjectType),
		FuncTemplateInformer: FuncTemplateInformer,
		FuncTemplateClient:   FuncTemplateClient,
		ReplicaSetClient:     ReplicaSetClient,
		ServiceClient:        ServiceClient,
		PodClient:            PodClient,
		queue:                cache.NewWorkQueue(),
	}

	_ = sc.FuncTemplateInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    sc.addFuncTemplate,
		UpdateFunc: sc.updateFuncTemplate,
		DeleteFunc: sc.deleteFuncTemplate,
	})

	return sc
}

type serverlessController struct {
	Kind string

	FuncTemplateInformer cache.Informer
	FuncTemplateClient   client.Interface
	ReplicaSetClient     client.Interface
	ServiceClient        client.Interface
	PodClient            client.Interface
	queue                cache.WorkQueue
}

func (sc *serverlessController) Run(ctx context.Context) {

	logger.ServerlessControllerLogger.Printf("[ServerlessController] start\n")

	go func() {
		defer logger.ServerlessControllerLogger.Printf("[ServerlessController] worker finish\n")
		sc.runWorker(ctx)
		// wait for controller manager stop
		<-ctx.Done()
	}()

	go func() {
		defer logger.ServerlessControllerLogger.Printf("[ServerlessController] periodicallyCheckFuncInstanceScale finish\n")
		sc.periodicallyCheckFuncInstanceScale(ctx)
		// wait for controller manager stop
		<-ctx.Done()
	}()

	return
}

func (sc *serverlessController) enqueueFunc(f *core.Func) {
	sc.queue.Enqueue(f)
	logger.ServerlessControllerLogger.Printf("enqueueFunc name %s\n", f.Spec.Name)
}

func (sc *serverlessController) processFuncStatusChange(oldFunc, curFunc *core.Func) error {
	// check if Counter in status change
	if oldFunc.Status.Counter != curFunc.Status.Counter {
		// modify rs replicas spec
		rsItem, err := sc.ReplicaSetClient.Get(curFunc.Status.ReplicaSetUID)
		if err != nil {
			return err
		}

		rs := rsItem.(*core.ReplicaSet)
		rs.Spec.Replicas = int32(curFunc.Status.Counter)

		// update rs Replicas number
		err = sc.updateRS(rs)
		if err != nil {
			return err
		}
	}
	return nil
}

func (sc *serverlessController) updateFuncTemplate(old, cur interface{}) {

	oldFunc := old.(*core.Func)
	curFunc := cur.(*core.Func)
	logger.ServerlessControllerLogger.Printf("Updating %s %s/%s\n", sc.Kind, curFunc.Namespace, curFunc.Name)

	if !reflect.DeepEqual(oldFunc.Status, curFunc.Status) {
		// deal with status change
		logger.ServerlessControllerLogger.Printf("status change: %s %s/%s\n", sc.Kind, curFunc.Namespace, curFunc.Name)
		err := sc.processFuncStatusChange(oldFunc, curFunc)
		if err != nil {
			logger.ServerlessControllerLogger.Printf("processFuncStatusChange err: %v\n", err)
		}
	}

	// if only status change, ignore
	if reflect.DeepEqual(oldFunc.Spec, curFunc.Spec) {
		logger.ServerlessControllerLogger.Printf("spec not change: %s %s/%s\n", sc.Kind, curFunc.Namespace, curFunc.Name)
		return
	}

	logger.ServerlessControllerLogger.Printf("spec change, delete old and create new: %s %s/%s\n", sc.Kind, curFunc.Namespace, curFunc.Name)

	// TODO: better update logic
	// delete service and rs of old template
	sc.deleteFuncTemplate(old)

	// recreate service and rs for new template
	sc.enqueueFunc(curFunc)
}

func (sc *serverlessController) addFuncTemplate(obj interface{}) {
	f := obj.(*core.Func)
	logger.ServerlessControllerLogger.Printf("Adding %s %s/%s\n", sc.Kind, f.Namespace, f.Name)
	sc.enqueueFunc(f)
}

func (sc *serverlessController) deleteFuncTemplate(obj interface{}) {
	f := obj.(*core.Func)

	logger.ServerlessControllerLogger.Printf("Deleting %s name %s\n", sc.Kind, f.Spec.Name)

	// delete service of current func template
	_, _, err := sc.ServiceClient.Delete(f.Status.ServiceUID)
	if err != nil {
		logger.ServerlessControllerLogger.Printf("Deleting %s name %s: delete its Service uid %v failed\n", sc.Kind, f.Spec.Name, f.Status.ServiceUID)
		err = nil
	}

	// delete rs of current func template, rs will be responsible for
	// delete all pods it is managing
	_, _, err = sc.ReplicaSetClient.Delete(f.Status.ReplicaSetUID)
	if err != nil {
		logger.ServerlessControllerLogger.Printf("Deleting %s name %s: delete its ReplicaSet uid %v failed\n", sc.Kind, f.Spec.Name, f.Status.ReplicaSetUID)
	}

}

const defaultWorkerSleepInterval = time.Duration(3) * time.Second

func (sc *serverlessController) runWorker(ctx context.Context) {
	//go wait.UntilWithContext(ctx, sc.worker, time.Second)
	for {
		select {
		case <-ctx.Done():
			logger.ServerlessControllerLogger.Printf("[worker] ctx.Done() received, worker of ServerlessController exit\n")
			return
		default:
			for sc.processNextWorkItem() {
			}
			time.Sleep(defaultWorkerSleepInterval)
		}
	}
}

func (sc *serverlessController) periodicallyCheckFuncInstanceScale(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			logger.ServerlessControllerLogger.Printf("[periodicallyCheckFuncInstanceScale] ctx.Done() received, worker of ServerlessController exit\n")
			return
		default:
			sc.checkFuncInstanceScale()
			time.Sleep(config.FuncInstanceScaleCheckInterval)
		}
	}
}

func (sc *serverlessController) checkFuncInstanceScale() {
	// Get all func template
	fList := sc.FuncTemplateInformer.List()
	for _, fItem := range fList {
		fTemplate := fItem.(*core.Func)
		if time.Since(fTemplate.Status.TimeStamp) > config.FuncInstanceScaleDownInterval {
			// time since last call has pass FuncInstanceScaleDownInterval, scale down
			currentInstanceNum := fTemplate.Status.Counter

			minInstanceNum := config.FuncDefaultMinInstanceNum
			if fTemplate.Spec.MinInstanceNum != nil {
				minInstanceNum = *fTemplate.Spec.MinInstanceNum
			}
			if currentInstanceNum <= minInstanceNum {
				// current instance num not enough, do not scale down
				continue
			}

			// scale down
			desiredInstanceNum := math.Min(float64(currentInstanceNum-config.FuncInstanceScaleDownNum), float64(minInstanceNum))

			fTemplate.Status.Counter = int(desiredInstanceNum)
			fTemplate.Status.TimeStamp = time.Now()

			err := sc.updateFuncStatus(fTemplate)
			if err != nil {
				logger.ServerlessControllerLogger.Printf("[checkFuncInstanceScale] scale down fail, err updateFuncStatus: %v\n", err)
			} else {
				logger.ServerlessControllerLogger.Printf("[checkFuncInstanceScale] scale down success for func %v, current instance num %v\n", fTemplate.Spec.Name, fTemplate.Status.Counter)
			}
		}
	}
}

func (sc *serverlessController) processNextWorkItem() bool {

	item, ok := sc.queue.Dequeue()
	if !ok {
		return false
	}

	serverlessFuncTemplate := item.(*core.Func)

	err := sc.processFuncCreate(serverlessFuncTemplate)
	if err != nil {
		logger.ServerlessControllerLogger.Printf("[processFuncCreate] err: %v\n", err)
		//enqueue if error happen when processing
		sc.queue.Enqueue(serverlessFuncTemplate)
		return false
	}

	return true
}

func (sc *serverlessController) immediatelyCreatePods(funcTemplate *core.Func, funcPodSpec core.PodSpec, funcPodLabels map[string]string, initReplicasNum int, funcTemplateOwnerRef meta.OwnerReference) error {

	// create initReplicasNum nums of pods immediately
	// pods will be manage by rs after rs created, don't be worry
	pod := generate.EmptyPod()
	pod.ObjectMeta = sc.generatePodObjectMetaForFunc(funcTemplate.Spec.Name, funcTemplateOwnerRef)
	pod.Spec = funcPodSpec
	pod.Labels = funcPodLabels

	num := 0
	for num < initReplicasNum {
		num += 1
		pod.Name = utils.AppendRandomNameSuffix(pod.Name)
		_, resp, err := sc.PodClient.Post(pod)
		if err != nil {
			logger.ServerlessControllerLogger.Printf("[immediatelyCreatePods] pod created immediately for func %v fail once\n", resp.UID, funcTemplate.Spec.Name)
			continue
		}
		logger.ServerlessControllerLogger.Printf("[immediatelyCreatePods] pod uid %v created immediately for func %v\n", resp.UID, funcTemplate.Spec.Name)
	}
	return nil
}

func (sc *serverlessController) generatePodSpecForFunc(funcTemplate *core.Func) (funcPodSpec core.PodSpec) {
	// TODO: @ wrj fill in pod spec
	funcPodSpec = core.PodSpec{
		Containers: []core.Container{{
			Name:            "gateway-" + funcTemplate.UID,
			Image:           "",  //TODO
			Env:             nil, //TODO
			ImagePullPolicy: core.PullIfNotPresent,
		}},
		RestartPolicy: core.RestartPolicyAlways,
	}
	return funcPodSpec
}

const (
	defaultCreateRetryTime = time.Second
	defaultRetryTimes      = 5
)

func (sc *serverlessController) createServiceForFunc(funcTemplate *core.Func, funcTemplateOwnerRef meta.OwnerReference, funcServiceLabels map[string]string) (uid types.UID, err error) {
	funcService := &core.Service{
		TypeMeta: meta.CreateTypeMeta(types.ServiceObjectType),
		ObjectMeta: meta.ObjectMeta{
			Name:      "gateway-" + funcTemplate.Name,
			Namespace: "serverless",
			OwnerReferences: []meta.OwnerReference{
				funcTemplateOwnerRef,
			},
		},
		Spec: core.ServiceSpec{ // TODO: @wjr add service spec
			Ports:     nil,
			Selector:  funcServiceLabels, // TODO: should label selector be this?
			ClusterIP: "",
			Type:      "",
		},
		Status: core.ServiceStatus{},
	}

	_, resp, err := sc.ServiceClient.Post(funcService)
	times := 0
	for err != nil {
		times += 1
		if times > defaultRetryTimes {
			logger.ServerlessControllerLogger.Printf("[createServiceForFunc] create svc failed, up to retry times limit\n")
			return meta.UIDNotGenerated, err
		}
		time.Sleep(defaultCreateRetryTime)
		logger.ServerlessControllerLogger.Printf("[createServiceForFunc] create svc failed, retry...\n")
		_, resp, err = sc.ServiceClient.Post(funcService)
	}
	logger.ServerlessControllerLogger.Printf("[createServiceForFunc] create svc for func %v success, svc uid %v\n", funcTemplate.Spec.Name, resp.UID)
	return resp.UID, nil
}

func (sc *serverlessController) generatePodObjectMetaForFunc(funcName string, funcTemplateOwnerRef meta.OwnerReference) meta.ObjectMeta {
	return meta.ObjectMeta{
		Name:      "funcTemplate-" + funcName + "-pod",
		Namespace: "serverless",
		OwnerReferences: []meta.OwnerReference{
			funcTemplateOwnerRef,
		},
	}
}

func (sc *serverlessController) createReplicaSetForFunc(funcTemplate *core.Func, funcReplicaSetLabels map[string]string, funcTemplateOwnerRef meta.OwnerReference, funcPodSpec core.PodSpec, funcPodLabels map[string]string, initReplicasNum int) (uid types.UID, err error) {

	// generate pod template from funcTemplate
	// match label of pod template and future service
	podTemplateSpec := core.PodTemplateSpec{
		ObjectMeta: sc.generatePodObjectMetaForFunc(funcTemplate.Spec.Name, funcTemplateOwnerRef),
		Spec:       funcPodSpec,
	}

	// add labels that match service and replica set
	podTemplateSpec.ObjectMeta.Labels = funcPodLabels

	// create replica set
	funcReplicaSet := &core.ReplicaSet{
		TypeMeta: meta.CreateTypeMeta(types.ReplicasetObjectType),
		ObjectMeta: meta.ObjectMeta{
			Name:      "func-" + funcTemplate.Spec.Name + "-rs",
			Namespace: "serverless",
			OwnerReferences: []meta.OwnerReference{
				funcTemplateOwnerRef,
			},
		},
		Spec: core.ReplicaSetSpec{
			Replicas: int32(initReplicasNum),
			Selector: meta.LabelSelector{
				MatchLabels: funcReplicaSetLabels, // add labels to match pod
			},
			Template: podTemplateSpec,
		},
		Status: core.ReplicaSetStatus{},
	}

	// add labels for rs itself, can be different from that of selector
	funcReplicaSet.ObjectMeta.Labels = funcReplicaSetLabels

	// create rs
	_, resp, err := sc.ReplicaSetClient.Post(funcReplicaSet)
	times := 0
	for err != nil {
		times += 1
		if times > defaultRetryTimes {
			logger.ServerlessControllerLogger.Printf("[createReplicaSetForFunc] create rs failed, up to retry times limit\n")
			return meta.UIDNotGenerated, err
		}
		time.Sleep(defaultCreateRetryTime)
		logger.ServerlessControllerLogger.Printf("[createReplicaSetForFunc] create rs failed, retry...\n")
		_, resp, err = sc.ReplicaSetClient.Post(funcReplicaSet)
	}
	logger.ServerlessControllerLogger.Printf("[createReplicaSetForFunc] create rs for func %v success, rs uid %v\n", funcTemplate.Spec.Name, resp.UID)
	return resp.UID, nil
}

func (sc *serverlessController) updateRS(rs *core.ReplicaSet) error {

	_, _, err := sc.ReplicaSetClient.Put(rs.UID, rs)
	times := 0
	for err != nil {
		times += 1
		if times > defaultRetryTimes {
			logger.ServerlessControllerLogger.Printf("[updateRS] updateRS failed, up to retry times limit\n")
			return err
		}
		time.Sleep(defaultCreateRetryTime)
		logger.ServerlessControllerLogger.Printf("[updateRS] updateRS failed, retry...\n")
		_, _, err = sc.ReplicaSetClient.Put(rs.UID, rs)
	}
	logger.ServerlessControllerLogger.Printf("[updateRS] updateRS for rs %v success\n", rs.UID)
	return nil
}

func (sc *serverlessController) updateFuncStatus(funcTemplate *core.Func) error {

	_, _, err := sc.FuncTemplateClient.Put(funcTemplate.Spec.Name, funcTemplate)
	times := 0
	for err != nil {
		times += 1
		if times > defaultRetryTimes {
			logger.ServerlessControllerLogger.Printf("[updateFuncStatus] updateFuncStatus failed, up to retry times limit\n")
			return err
		}
		time.Sleep(defaultCreateRetryTime)
		logger.ServerlessControllerLogger.Printf("[updateFuncStatus] updateFuncStatus failed, retry...\n")
		_, _, err = sc.FuncTemplateClient.Put(funcTemplate.Spec.Name, funcTemplate)
	}
	logger.ServerlessControllerLogger.Printf("[updateFuncStatus] updateFuncStatus for func %v success\n", funcTemplate.Spec.Name)
	return nil
}

// process func template create and spec update event
func (sc *serverlessController) processFuncCreate(funcTemplate *core.Func) error {

	funcTemplate.Name = funcTemplate.Spec.Name

	funcTemplateOwnerRef := funcTemplate.GenerateOwnerReference()

	// labels that match pod template for rs
	funcReplicaSetLabels := map[string]string{
		"_serverless_replicaset": funcTemplate.Name,
	}

	// labels that match pod for service
	funcServiceLabels := map[string]string{
		"_gateway": funcTemplate.Name,
	}

	// labels for pod
	funcPodLabels := map[string]string{
		"_gateway":               funcTemplate.Name,
		"_serverless_replicaset": funcTemplate.Name,
	}

	// generate pod spec for func
	funcPodSpec := sc.generatePodSpecForFunc(funcTemplate)

	// immediately create pods if initReplicasNum is not zero
	initReplicasNum := config.FuncDefaultInitInstanceNum
	if funcTemplate.Spec.InitInstanceNum != nil {
		initReplicasNum = *funcTemplate.Spec.InitInstanceNum
		err := sc.immediatelyCreatePods(funcTemplate, funcPodSpec, funcPodLabels, initReplicasNum, funcTemplateOwnerRef)
		if err != nil {
			return err
		}
	}

	// create service
	svcUID, err := sc.createServiceForFunc(funcTemplate, funcTemplateOwnerRef, funcServiceLabels)
	if err != nil {
		return err
	}

	// immediately pods may be running now, func call may success here!

	// create replica set
	rsUID, err := sc.createReplicaSetForFunc(funcTemplate, funcReplicaSetLabels, funcTemplateOwnerRef, funcPodSpec, funcPodLabels, initReplicasNum)
	if err != nil {
		return err
	}

	// modify func status
	funcTemplate.Status.ReplicaSetUID = rsUID
	funcTemplate.Status.ServiceUID = svcUID
	funcTemplate.Status.Counter = initReplicasNum
	funcTemplate.Status.TimeStamp = time.Now() // init timestamp by current time

	// update func template
	err = sc.updateFuncStatus(funcTemplate)
	if err != nil {
		return err
	}

	return nil
}

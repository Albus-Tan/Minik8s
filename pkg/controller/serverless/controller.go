//package serverless
//
//import (
//	"context"
//	"minik8s/pkg/api/core"
//	"minik8s/pkg/api/generate"
//	"minik8s/pkg/api/meta"
//	"minik8s/pkg/api/types"
//	client "minik8s/pkg/apiclient/interface"
//	"minik8s/pkg/controller/cache"
//	"minik8s/pkg/logger"
//	"time"
//)
//
//type ServerlessController interface {
//	Run(ctx context.Context)
//}
//
//func NewServerlessController(
//	ServerlessInformer cache.Informer,
//	ServerlessClient client.Interface,
//	ReplicaSetInformer cache.Informer,
//	ReplicaSetClient client.Interface,
//	ServiceClient client.Interface,
//) ServerlessController {
//
//	serverlessc := &serverlessController{
//		Kind:               string(types.FuncTemplateObjectType),
//		ServerlessInformer: ServerlessInformer,
//		ServerlessClient:   ServerlessClient,
//		ReplicaSetInformer: ReplicaSetInformer,
//		ReplicaSetClient:   ReplicaSetClient,
//		ServiceClient:      ServiceClient,
//
//		queue: cache.NewWorkQueue(),
//	}
//
//	_ = serverlessc.ServerlessInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
//		AddFunc:    serverlessc.addFunc,
//		UpdateFunc: serverlessc.updateFunc
//		DeleteFunc: serverlessc.deleteFunc,
//	})
//
//	return serverlessc
//}
//
//type serverlessController struct {
//	Kind string
//
//	ServerlessInformer cache.Informer
//	ServerlessClient   client.Interface
//	ReplicaSetInformer cache.Informer
//	ReplicaSetClient   client.Interface
//	ServiceClient      client.Interface
//	queue              cache.WorkQueue
//}
//
//func (serverlessc *serverlessController) Run(ctx context.Context) {
//
//	go func() {
//		logger.ServerlessLogger.Printf("[ServerlessController] start\n")
//		defer logger.ServerlessLogger.Printf("[ServerlessController] finish\n")
//
//		serverlessc.runWorker(ctx)
//
//		//wait for controller manager stop
//		<-ctx.Done()
//	}()
//	return
//}
//
//func (serverlessc *serverlessController) FuncKeyFunc(Func *core.Func) string {
//	return Func.GetUID()
//}
//
//func (serverlessc *serverlessController) enqueueFunc(Func *core.Func) {
//	key := serverlessc.FuncKeyFunc(Func)
//	serverlessc.queue.Enqueue(Func)
//	logger.ServerlessLogger.Printf("enqueueFunc uid %s\n", key)
//}
//
//func (serverlessc *serverlessController) updateFunc(old, cur interface{}) {
//	Func := obj.(*core.Func)
//	logger.ServerlessLogger.Printf("Adding %s %s/%s\n", serverlessc.Kind, Func.Namespace, Func.Name)
//	serverlessc.enqueueFunc(Func)
//}
//
//func (serverlessc *serverlessController) addFunc(obj interface{}) {
//	Func := obj.(*core.Func)
//	logger.ServerlessLogger.Printf("Adding %s %s/%s\n", serverlessc.Kind, Func.Namespace, Func.Name)
//	serverlessc.enqueueFunc(Func)
//}
//
//func (serverlessc *serverlessController) deleteFunc(obj interface{}) {
//	Func := obj.(*core.Func)
//
//	logger.ServerlessLogger.Printf("Deleting %s, uid %s\n", serverlessc.Kind, Func.UID)
//
//	_, _, err := serverlessc.ServiceClient.Delete(Func.Status.ServiceUID)
//	if err != nil {
//		return
//	}
//	_, _, err = serverlessc.PodClient.Delete(Func.Status.PodUID)
//	if err != nil {
//		return
//	}
//
//}
//
//const defaultWorkeFuncleepInterval = time.Duration(3) * time.Second
//
//func (serverlessc *serverlessController) runWorker(ctx context.Context) {
//	//go wait.UntilWithContext(ctx, serverlessc.worker, time.Second)
//	for {
//		select {
//		case <-ctx.Done():
//			logger.ServerlessLogger.Printf("[worker] ctx.Done() received, worker of ServerlessController exit\n")
//			return
//		default:
//			for serverlessc.processNextWorkItem() {
//			}
//			time.Sleep(defaultWorkeFuncleepInterval)
//		}
//	}
//}
//
//func (serverlessc *serverlessController) processNextWorkItem() bool {
//
//	item, ok := serverlessc.queue.Dequeue()
//	if !ok {
//		return false
//	}
//
//	serverless := item.(*core.Func)
//
//	err := serverlessc.processFuncCreate(serverless)
//	if err != nil {
//		logger.ServerlessLogger.Printf("[processFuncCreate] err: %v\n", err)
//		//enqueue if error happen when processing
//		serverlessc.queue.Enqueue(serverless)
//		return false
//	}
//
//	return true
//}
//
//func (serverlessc *serverlessController) processFuncCreate(serverless *core.Func) error {
//
//	//TODO
//	//	process serverless create event
//
//	pod := generate.EmptyPod()
//	pod.Name = "serverless-" + serverless.Name
//	pod.Spec = core.PodSpec{
//		Containers: []core.Container{{
//			Name:            "gateway-" + serverless.UID,
//			Image:           "",  //TODO
//			Env:             nil, //TODO
//			ImagePullPolicy: core.PullIfNotPresent,
//		}},
//		RestartPolicy: core.RestartPolicyAlways,
//	}
//	pod.Labels = map[string]string{
//		"_gateway": serverless.Name,
//	}
//	_, pr, err := serverlessc.PodClient.Post(pod)
//	if err != nil {
//		return err
//	}
//	svc := &core.Service{
//		TypeMeta: meta.CreateTypeMeta(types.ServiceObjectType),
//		ObjectMeta: meta.ObjectMeta{
//			Name: "gateway-" + serverless.Name,
//		},
//		Spec: core.ServiceSpec{
//			Ports:     nil,
//			Selector:  nil,
//			ClusterIP: "",
//			Type:      "",
//		},
//		Status: core.ServiceStatus{},
//	}
//	_, sr, err := serverlessc.ServiceClient.Post(svc)
//	if err != nil {
//		return err
//	}
//
//	serverless.Status.PodUID = pr.UID
//	serverless.Status.ServiceUID = sr.UID
//	_, _, err = serverlessc.ServerlessClient.Put(serverless.UID, serverless)
//	if err != nil {
//		return err
//	}
//	//TODO delete service use uid saved in status
//	//TODO delete pod use uid saved in status
//	return nil
//}

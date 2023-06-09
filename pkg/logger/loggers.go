package logger

import "minik8s/utils"

type Logger utils.ComponentLogger

var ApiServerLogger Logger
var ApiClientLogger Logger
var ControllerManagerLogger Logger
var ReplicaSetControllerLogger Logger
var HorizontalControllerLogger Logger
var SchedulerLogger Logger
var GpuServerLogger Logger
var KubectlLogger Logger
var KubeletLogger Logger
var DNSControllerLogger Logger
var ServerlessControllerLogger Logger
var PodControllerLogger Logger

func init() {
	ApiServerLogger = utils.NewComponentLogger("ApiServer")
	ApiClientLogger = utils.NewComponentLogger("ApiClient")
	ControllerManagerLogger = utils.NewComponentLogger("ControllerManager")
	ReplicaSetControllerLogger = utils.NewComponentLogger("ReplicaSetController")
	HorizontalControllerLogger = utils.NewComponentLogger("HorizontalController")
	SchedulerLogger = utils.NewComponentLogger("Scheduler")
	KubectlLogger = utils.NewComponentLogger("Kubectl")
	KubeletLogger = utils.NewComponentLogger("Kubelet")
	GpuServerLogger = utils.NewComponentLogger("GpuServer")
	DNSControllerLogger = utils.NewComponentLogger("DNSController")
	ServerlessControllerLogger = utils.NewComponentLogger("ServerlessController")
	PodControllerLogger = utils.NewComponentLogger("PodController")
}

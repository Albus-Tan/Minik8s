package logger

import "minik8s/utils"

type Logger utils.ComponentLogger

var ApiServerLogger Logger
var ApiClientLogger Logger
var ControllerManagerLogger Logger
var ReplicaSetControllerLogger Logger
var KubectlLogger Logger
var KubeletLogger Logger

func init() {
	ApiServerLogger = utils.NewComponentLogger("ApiServer")
	ApiClientLogger = utils.NewComponentLogger("ApiClient")
	ControllerManagerLogger = utils.NewComponentLogger("ControllerManager")
	ReplicaSetControllerLogger = utils.NewComponentLogger("ReplicaSetController")
	KubectlLogger = utils.NewComponentLogger("Kubectl")
	KubeletLogger = utils.NewComponentLogger("Kubelet")
}
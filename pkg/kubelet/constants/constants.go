package constants

import "minik8s/pkg/api/core"

var InitialPauseContainer = core.Container{
	Name:            "pause",
	Image:           "kubernetes/pause:latest",
	ImagePullPolicy: "IfNotPresent",
	//Resources: map[string]string{
	//	"cpu":    "2",
	//	"memory": "256MB",
	//},
}

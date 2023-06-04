package kubectl

import (
	"errors"
	"fmt"
	"minik8s/pkg/api/types"
)

func ParseType(ty string) (types.ApiObjectType, error) {
	switch ty {
	case "pod", "pods":
		return types.PodObjectType, nil
	case "service", "svc", "services":
		return types.ServiceObjectType, nil
	case "nodes", "node":
		return types.NodeObjectType, nil
	case "replicaset", "rs", "replicasets":
		return types.ReplicasetObjectType, nil
	case "hpa", "hpas":
		return types.HorizontalPodAutoscalerObjectType, nil
	case "func", "f", "funcs":
		return types.FuncTemplateObjectType, nil
	case "job", "j", "jobs":
		return types.JobObjectType, nil
	case "dns":
		return types.DnsObjectType, nil
	default:
		errMsg := fmt.Sprintf("No ObjectType %v", ty)
		return types.ErrorObjectType, errors.New(errMsg)
	}
}

package generate

import (
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/types"
	"minik8s/utils"
)

func PodFromReplicaSet(rs *core.ReplicaSet) *core.Pod {
	podTemplate := rs.Spec.Template
	newPod := &core.Pod{
		TypeMeta:   meta.CreateTypeMeta(types.PodObjectType),
		ObjectMeta: podTemplate.ObjectMeta,
		Spec:       podTemplate.Spec,
		Status:     core.PodStatus{},
	}
	newPod.UID = meta.UIDNotGenerated
	newPod.Name = utils.AppendRandomNameSuffix(rs.Name)
	return newPod
}

func EmptyPod() *core.Pod {
	return &core.Pod{
		TypeMeta:   meta.CreateTypeMeta(types.PodObjectType),
		ObjectMeta: meta.ObjectMeta{},
		Spec:       core.PodSpec{},
		Status:     core.PodStatus{},
	}
}

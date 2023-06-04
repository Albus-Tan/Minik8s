package core

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/types"
	"strconv"
)

// ReplicaSet ensures that a specified number of pod replicas are running at any given time.
type ReplicaSet struct {
	meta.TypeMeta `json:",inline"`

	// If the Labels of a ReplicaSet are empty, they are defaulted to
	// be the same as the Pod(s) that the ReplicaSet manages.
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	meta.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines the specification of the desired behavior of the ReplicaSet.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec ReplicaSetSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Status is the most recently observed status of the ReplicaSet.
	// This data may be out of date by some window of time.
	// Populated by the system.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Status ReplicaSetStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

func (r *ReplicaSet) PrintBrief() {
	fmt.Printf("%-20s\t%-40s\t%-15s\t%-15s\n", "NAME", "UID", "DESIRED", "CURRENT")
	fmt.Printf("%-20s\t%-40s\t%-15d\t%-15d\n", r.Name, r.UID, r.Spec.Replicas, r.Status.Replicas)
}

func (r *ReplicaSet) DeleteOwnerReference(uid types.UID) {
	has := false
	idx := 0
	for i, o := range r.OwnerReferences {
		if o.UID == uid {
			has = true
			idx = i
			break
		}
	}
	if has {
		r.OwnerReferences = append(r.OwnerReferences[:idx], r.OwnerReferences[idx+1:]...)
	}
}

func (r *ReplicaSet) AppendOwnerReference(reference meta.OwnerReference) {
	r.OwnerReferences = append(r.OwnerReferences, reference)
}

func (r *ReplicaSet) GenerateOwnerReference() meta.OwnerReference {
	return meta.OwnerReference{
		APIVersion: r.APIVersion,
		Kind:       r.Kind,
		Name:       r.Name,
		UID:        r.UID,
		Controller: false,
	}
}

func (r *ReplicaSet) SetUID(uid types.UID) {
	r.ObjectMeta.UID = uid
}

func (r *ReplicaSet) GetUID() types.UID {
	return r.ObjectMeta.UID
}

func (r *ReplicaSet) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &r)
}

func (r *ReplicaSet) JsonMarshal() ([]byte, error) {
	return json.Marshal(r)
}

func (r *ReplicaSet) JsonUnmarshalStatus(data []byte) error {
	return json.Unmarshal(data, &(r.Status))
}

func (r *ReplicaSet) JsonMarshalStatus() ([]byte, error) {
	return json.Marshal(r.Status)
}

func (r *ReplicaSet) SetStatus(s IApiObjectStatus) bool {
	status, ok := s.(*ReplicaSetStatus)
	if ok {
		r.Status = *status
	}
	return ok
}

func (r *ReplicaSet) GetStatus() IApiObjectStatus {
	return &r.Status
}

func (r *ReplicaSet) GetResourceVersion() string {
	return r.ObjectMeta.ResourceVersion
}

func (r *ReplicaSet) SetResourceVersion(version string) {
	r.ObjectMeta.ResourceVersion = version
}

func (r *ReplicaSet) CreateFromEtcdString(str string) error {
	return r.JsonUnmarshal([]byte(str))
}

// ReplicaSetSpec is the specification of a ReplicaSet.
type ReplicaSetSpec struct {
	// Replicas is the number of desired replicas.
	// This is a pointer to distinguish between explicit zero and unspecified.
	// Defaults to 1.
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/replicationcontroller/#what-is-a-replicationcontroller
	// +optional
	Replicas int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// Selector is a label query over pods that should match the replica count.
	// Label keys and values that must match in order to be controlled by this replica set.
	// It must match the pod template's labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	Selector meta.LabelSelector `json:"selector" protobuf:"bytes,2,opt,name=selector"`

	// Template is the object that describes the pod that will be created if
	// insufficient replicas are detected.
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/replicationcontroller#pod-template
	// +optional
	Template PodTemplateSpec `json:"template,omitempty" protobuf:"bytes,3,opt,name=template"`
}

// ReplicaSetStatus represents the current status of a ReplicaSet.
type ReplicaSetStatus struct {
	// Replicas is the most recently observed number of replicas.
	// More info: https://kubernetes.io/docs/concepts/workloads/controllers/replicationcontroller/#what-is-a-replicationcontroller
	Replicas int32 `json:"replicas" protobuf:"varint,1,opt,name=replicas"`

	// The number of pods that have labels matching the labels of the pod template of the replicaset.
	// +optional
	FullyLabeledReplicas int32 `json:"fullyLabeledReplicas,omitempty" protobuf:"varint,2,opt,name=fullyLabeledReplicas"`

	// readyReplicas is the number of pods targeted by this ReplicaSet with a Ready Condition.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty" protobuf:"varint,4,opt,name=readyReplicas"`

	// The number of available replicas (ready for at least minReadySeconds) for this replica set.
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty" protobuf:"varint,5,opt,name=availableReplicas"`

	// ObservedGeneration reflects the generation of the most recently observed ReplicaSet.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,3,opt,name=observedGeneration"`

	// Represents the latest available observations of a replica set's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []ReplicaSetCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,6,rep,name=conditions"`
}

func (r *ReplicaSetStatus) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &r)
}

func (r *ReplicaSetStatus) JsonMarshal() ([]byte, error) {
	return json.Marshal(r)
}

type ReplicaSetConditionType string

// These are valid conditions of a replica set.
const (
	// ReplicaSetReplicaFailure is added in a replica set when one of its pods fails to be created
	// due to insufficient quota, limit ranges, pod security policy, node selectors, etc. or deleted
	// due to kubelet being down or finalizers are failing.
	ReplicaSetReplicaFailure ReplicaSetConditionType = "ReplicaFailure"
)

// ReplicaSetCondition describes the state of a replica set at a certain point.
type ReplicaSetCondition struct {
	// Type of replica set condition.
	Type ReplicaSetConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=ReplicaSetConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// The last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime types.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
}

type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in the condition.
// "ConditionFalse" means a resource is not in the condition. "ConditionUnknown" means kubernetes
// can't decide if a resource is in the condition or not. In the future, we could add other
// intermediate conditions, e.g. ConditionDegraded.
const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

type ReplicaSetList struct {
	meta.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	meta.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// List of pods.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md
	Items []ReplicaSet `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func (r *ReplicaSetList) PrintBrief() {
	fmt.Printf("%-20s\t%-40s\t%-15s\t%-15s\n", "NAME", "UID", "DESIRED", "CURRENT")
	for _, item := range r.Items {
		fmt.Printf("%-20s\t%-40s\t%-15d\t%-15d\n", item.Name, item.UID, item.Spec.Replicas, item.Status.Replicas)
	}
}

func (r *ReplicaSetList) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &r)
}

func (r *ReplicaSetList) JsonMarshal() ([]byte, error) {
	return json.Marshal(r)
}

func (r *ReplicaSetList) AddItemFromStr(objectStr string) error {
	object := &ReplicaSet{}
	buf, err := strconv.Unquote(objectStr)
	err = object.JsonUnmarshal([]byte(buf))
	if err != nil {
		return err
	}
	r.Items = append(r.Items, *object)
	return nil
}

func (r *ReplicaSetList) AppendItemsFromStr(objectStrs []string) error {
	for _, obj := range objectStrs {
		object := &ReplicaSet{}
		err := object.JsonUnmarshal([]byte(obj))
		if err != nil {
			return err
		}
		r.Items = append(r.Items, *object)
	}
	return nil
}

func (r *ReplicaSetList) GetItems() any {
	return r.Items
}

func (r *ReplicaSetList) GetIApiObjectArr() (res []IApiObject) {
	for _, item := range r.Items {
		itemTemp := item
		res = append(res, &itemTemp)
	}
	return res
}

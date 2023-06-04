package core

import (
	"encoding/json"
	"fmt"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/types"
	"strconv"
)

// Pod is a collection of containers that can run on a host. This resource is created
// by clients and scheduled onto hosts.
type Pod struct {
	meta.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	meta.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the pod.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec PodSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Most recently observed status of the pod.
	// This data may not be up-to-date.
	// Populated by the system.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Status PodStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

func (p *Pod) PrintBrief() {
	fmt.Printf("%-20s\t%-40s\t%-8s\t%-15s\t%-15s\n", "NAME", "UID", "NODE", "STATUS", "IP")
	fmt.Printf("%-20s\t%-40s\t%-8s\t%-15s\t%-15s\n", p.Name, p.UID, p.Spec.NodeName, p.Status.Phase, p.Status.PodIP)
}

func (p *Pod) DeleteOwnerReference(uid types.UID) {
	has := false
	idx := 0
	for i, o := range p.OwnerReferences {
		if o.UID == uid {
			has = true
			idx = i
			break
		}
	}
	if has {
		p.OwnerReferences = append(p.OwnerReferences[:idx], p.OwnerReferences[idx+1:]...)
	}
}

func (p *Pod) AppendOwnerReference(reference meta.OwnerReference) {
	p.OwnerReferences = append(p.OwnerReferences, reference)
}

func (p *Pod) GenerateOwnerReference() meta.OwnerReference {
	return meta.OwnerReference{
		APIVersion: p.APIVersion,
		Kind:       p.Kind,
		Name:       p.Name,
		UID:        p.UID,
		Controller: false,
	}
}

func (p *Pod) CreateFromEtcdString(str string) error {
	return p.JsonUnmarshal([]byte(str))
}

func (p *Pod) SetResourceVersion(version string) {
	p.ObjectMeta.ResourceVersion = version
}

func (p *Pod) GetResourceVersion() string {
	return p.ObjectMeta.ResourceVersion
}

func (p *Pod) SetStatus(s IApiObjectStatus) bool {
	status, ok := s.(*PodStatus)
	if ok {
		p.Status = *status
	}
	return ok
}

func (p *Pod) GetStatus() IApiObjectStatus {
	return &p.Status
}

func (p *Pod) JsonUnmarshalStatus(data []byte) error {
	return json.Unmarshal(data, &(p.Status))
}

func (p *Pod) JsonMarshalStatus() ([]byte, error) {
	return json.Marshal(p.Status)
}

func (p *Pod) JsonMarshal() ([]byte, error) {
	return json.Marshal(p)
}

func (p *Pod) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &p)
}

func (p *Pod) SetUID(uid types.UID) {
	p.ObjectMeta.UID = uid
}

func (p *Pod) GetUID() types.UID {
	return p.ObjectMeta.UID
}

// PodSpec is a description of a pod.
type PodSpec struct {
	// List of volumes that can be mounted by containers belonging to the pod.
	// More info: https://kubernetes.io/docs/concepts/storage/volumes
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Volumes []Volume `json:"volumes,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name" protobuf:"bytes,1,rep,name=volumes"`
	// List of initialization containers belonging to the pod.
	// Init containers are executed in order prior to containers being started. If any
	// init container fails, the pod is considered to have failed and is handled according
	// to its restartPolicy. The name for an init container or normal container must be
	// unique among all containers.
	// Init containers may not have Lifecycle actions, Readiness probes, Liveness probes, or Startup probes.
	// The resourceRequirements of an init container are taken into account during scheduling
	// by finding the highest request/limit for each resource type, and then using the max of
	// that value or the sum of the normal containers. Limits are applied to init containers
	// in a similar fashion.
	// Init containers cannot currently be added or removed.
	// Cannot be updated.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/init-containers/
	// +patchMergeKey=name
	// +patchStrategy=merge
	InitContainers []Container `json:"initContainers,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,20,rep,name=initContainers"`
	// List of containers belonging to the pod.
	// Containers cannot currently be added or removed.
	// There must be at least one container in a Pod.
	// Cannot be updated.
	// +patchMergeKey=name
	// +patchStrategy=merge
	Containers []Container `json:"containers" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,2,rep,name=containers"`
	// Restart policy for all containers within the pod.
	// One of Always, OnFailure, Never. In some contexts, only a subset of those values may be permitted.
	// Default to Always.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#restart-policy
	// +optional
	RestartPolicy RestartPolicy `json:"restartPolicy,omitempty" protobuf:"bytes,3,opt,name=restartPolicy,casttype=RestartPolicy"`

	// NodeName is a request to schedule this pod onto a specific node. If it is non-empty,
	// the scheduler simply schedules this pod onto that node, assuming that it fits resource
	// requirements.
	// +optional
	NodeName string `json:"nodeName,omitempty" protobuf:"bytes,10,opt,name=nodeName"`

	// TODO
	ExposedPorts []string `json:"exposedPorts,omitempty"`

	// TODO
	BindPorts map[string]string `json:"bindPorts,omitempty"`

	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *Affinity `json:"affinity,omitempty" protobuf:"bytes,18,opt,name=affinity"`
}

// RestartPolicy describes how the container should be restarted.
// Only one of the following restart policies may be specified.
// If none of the following policies is specified, the default one
// is RestartPolicyAlways.
// +enum
type RestartPolicy string

const (
	RestartPolicyAlways    RestartPolicy = "Always"
	RestartPolicyOnFailure RestartPolicy = "OnFailure"
	RestartPolicyNever     RestartPolicy = "Never"
)

// PodStatus represents information about the status of a pod. Status may trail the actual
// state of a system, especially if the node that hosts the pod cannot contact the control
// plane.
type PodStatus struct {
	// The phase of a Pod is a simple, high-level summary of where the Pod is in its lifecycle.
	// The conditions array, the reason and message fields, and the individual container status
	// arrays contain more detail about the pod's status.
	// There are five possible phase values:
	//
	// Pending: The pod has been accepted by the Kubernetes system, but one or more of the
	// container images has not been created. This includes time before being scheduled as
	// well as time spent downloading images over the network, which could take a while.
	// Running: The pod has been bound to a node, and all the containers have been created.
	// At least one container is still running, or is in the process of starting or restarting.
	// Succeeded: All containers in the pod have terminated in success, and will not be restarted.
	// Failed: All containers in the pod have terminated, and at least one container has
	// terminated in failure. The container either exited with non-zero status or was terminated
	// by the system.
	// Unknown: For some reason the state of the pod could not be obtained, typically due to an
	// error in communicating with the host of the pod.
	//
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#pod-phase
	// +optional
	Phase PodPhase `json:"phase,omitempty" protobuf:"bytes,1,opt,name=phase,casttype=PodPhase"`

	// IP address of the host to which the pod is assigned. Empty if not yet scheduled.
	// +optional
	HostIP string `json:"hostIP,omitempty" protobuf:"bytes,5,opt,name=hostIP"`
	// IP address allocated to the pod. Routable at least within the cluster.
	// Empty if not yet allocated.
	// +optional
	PodIP string `json:"podIP,omitempty" protobuf:"bytes,6,opt,name=podIP"`

	// The list has one entry per container in the manifest.
	// More info: https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle#pod-and-container-status
	// +optional
	ContainerStatuses []ContainerStatus `json:"containerStatuses,omitempty" protobuf:"bytes,8,rep,name=containerStatuses"`
}

func DefaultPosStatus() PodStatus {
	return PodStatus{
		Phase:             PodPending,
		HostIP:            "unknown",
		PodIP:             "unknown",
		ContainerStatuses: nil,
	}
}

func (p *PodStatus) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &p)
}

func (p *PodStatus) JsonMarshal() ([]byte, error) {
	return json.Marshal(p)
}

// PodPhase is a label for the condition of a pod at the current time.
// +enum
type PodPhase string

// These are the valid statuses of pods.
const (
	// PodPending means the pod has been accepted by the system, but one or more of the containers
	// has not been started. This includes time before being bound to a node, as well as time spent
	// pulling images onto the host.
	PodPending PodPhase = "Pending"
	// PodRunning means the pod has been bound to a node and all the containers have been started.
	// At least one container is still running or is in the process of being restarted.
	PodRunning PodPhase = "Running"
	// PodSucceeded means that all containers in the pod have voluntarily terminated
	// with a container exit code of 0, and the system is not going to restart any of these containers.
	PodSucceeded PodPhase = "Succeeded"
	// PodFailed means that all containers in the pod have terminated, and at least one container has
	// terminated in a failure (exited with a non-zero exit code or was stopped by the system).
	PodFailed PodPhase = "Failed"
	// PodUnknown means that for some reason the state of the pod could not be obtained, typically due
	// to an error in communicating with the host of the pod.
	// Deprecated: It isn't being set since 2015 (74da3b14b0c0f658b3bb8d2def5094686d0e9095)
	PodUnknown PodPhase = "Unknown"
)

// PodTemplateSpec describes the data a pod should have when created from a template
type PodTemplateSpec struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	meta.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the pod.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec PodSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// PodList is a list of Pods.
type PodList struct {
	meta.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	meta.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// List of pods.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md
	Items []Pod `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func (p *PodList) PrintBrief() {
	fmt.Printf("%-20s\t%-40s\t%-8s\t%-15s\t%-15s\n", "NAME", "UID", "NODE", "STATUS", "IP")
	for _, item := range p.Items {
		fmt.Printf("%-20s\t%-40s\t%-8s\t%-15s\t%-15s\n", item.Name, item.UID, item.Spec.NodeName, item.Status.Phase, item.Status.PodIP)
	}
}

func (p *PodList) AppendItemsFromStr(objectStrs []string) error {
	for _, obj := range objectStrs {
		object := &Pod{}
		err := object.JsonUnmarshal([]byte(obj))
		if err != nil {
			return err
		}
		p.Items = append(p.Items, *object)
	}
	return nil
}

func (p *PodList) AddItemFromStr(objectStr string) error {
	object := &Pod{}
	buf, err := strconv.Unquote(objectStr)
	err = object.JsonUnmarshal([]byte(buf))
	if err != nil {
		return err
	}
	p.Items = append(p.Items, *object)
	return nil
}

func (p *PodList) GetItems() any {
	return p.Items
}

func (p *PodList) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &p)
}

func (p *PodList) JsonMarshal() ([]byte, error) {
	return json.Marshal(p)
}

func (p *PodList) GetIApiObjectArr() (res []IApiObject) {
	for _, item := range p.Items {
		itemTemp := item
		res = append(res, &itemTemp)
	}
	return res
}

// Affinity is a group of affinity scheduling rules.
type Affinity struct {
	// Describes pod anti-affinity scheduling rules (e.g. avoid putting this pod in the same node, zone, etc. as some other pod(s)).
	// +optional
	PodAntiAffinity PodAntiAffinity `json:"podAntiAffinity,omitempty" protobuf:"bytes,3,opt,name=podAntiAffinity"`
}

// PodAntiAffinity Pod anti affinity is a group of inter pod anti affinity scheduling rules.
type PodAntiAffinity struct {
	// If the anti-affinity requirements specified by this field are not met at
	// scheduling time, the pod will not be scheduled onto the node.
	// If the anti-affinity requirements specified by this field cease to be met
	// at some point during pod execution (e.g. due to a pod label update), the
	// system may or may not try to eventually evict the pod from its node.
	// When there are multiple elements, the lists of nodes corresponding to each
	// podAffinityTerm are intersected, i.e. all terms must be satisfied.
	// +optional
	RequiredDuringSchedulingIgnoredDuringExecution []PodAffinityTerm `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty" protobuf:"bytes,1,rep,name=requiredDuringSchedulingIgnoredDuringExecution"`
}

type PodAffinityTerm struct {
	// A label query over a set of resources, in this case pods.
	// +optional
	LabelSelector *meta.LabelSelector `json:"labelSelector,omitempty" protobuf:"bytes,1,opt,name=labelSelector"`
	// This pod should be co-located (affinity) or not co-located (anti-affinity) with the pods matching
	// the labelSelector in the specified namespaces, where co-located is defined as running on a node
	// whose value of the label with key topologyKey matches that of any node on which any of the
	// selected pods is running.
	// Empty topologyKey is not allowed.
	TopologyKey string `json:"topologyKey" protobuf:"bytes,3,opt,name=topologyKey"`
}

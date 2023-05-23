package core

import (
	"encoding/json"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/types"
	"strconv"
)

// HorizontalPodAutoscaler is the configuration for a horizontal pod
// autoscaler, which automatically manages the replica count of any resource
// implementing the scale subresource based on the metrics specified.
type HorizontalPodAutoscaler struct {
	meta.TypeMeta `json:",inline"`
	// Metadata is the standard object metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	meta.ObjectMeta `json:"metadata,omitempty"`

	// spec is the specification for the behaviour of the autoscaler.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status.
	// +optional
	Spec HorizontalPodAutoscalerSpec `json:"spec,omitempty"`

	// status is the current information about the autoscaler.
	// +optional
	Status HorizontalPodAutoscalerStatus `json:"status,omitempty"`
}

func (h *HorizontalPodAutoscaler) SetUID(uid types.UID) {
	h.ObjectMeta.UID = uid
}

func (h *HorizontalPodAutoscaler) GetUID() types.UID {
	return h.ObjectMeta.UID
}

func (h *HorizontalPodAutoscaler) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &h)
}

func (h *HorizontalPodAutoscaler) JsonMarshal() ([]byte, error) {
	return json.Marshal(h)
}

func (h *HorizontalPodAutoscaler) JsonUnmarshalStatus(data []byte) error {
	return json.Unmarshal(data, &(h.Status))
}

func (h *HorizontalPodAutoscaler) JsonMarshalStatus() ([]byte, error) {
	return json.Marshal(h.Status)
}

func (h *HorizontalPodAutoscaler) SetStatus(s IApiObjectStatus) bool {
	status, ok := s.(*HorizontalPodAutoscalerStatus)
	if ok {
		h.Status = *status
	}
	return ok
}

func (h *HorizontalPodAutoscaler) GetStatus() IApiObjectStatus {
	return &h.Status
}

func (h *HorizontalPodAutoscaler) GetResourceVersion() string {
	return h.ObjectMeta.ResourceVersion
}

func (h *HorizontalPodAutoscaler) SetResourceVersion(version string) {
	h.ObjectMeta.ResourceVersion = version
}

func (h *HorizontalPodAutoscaler) CreateFromEtcdString(str string) error {
	return h.JsonUnmarshal([]byte(str))
}

func (h *HorizontalPodAutoscaler) GenerateOwnerReference() meta.OwnerReference {
	return meta.OwnerReference{
		APIVersion: h.APIVersion,
		Kind:       h.Kind,
		Name:       h.Name,
		UID:        h.UID,
		Controller: false,
	}
}

func (h *HorizontalPodAutoscaler) AppendOwnerReference(reference meta.OwnerReference) {
	h.OwnerReferences = append(h.OwnerReferences, reference)
}

func (h *HorizontalPodAutoscaler) DeleteOwnerReference(uid types.UID) {
	has := false
	idx := 0
	for i, o := range h.OwnerReferences {
		if o.UID == uid {
			has = true
			idx = i
			break
		}
	}
	if has {
		h.OwnerReferences = append(h.OwnerReferences[:idx], h.OwnerReferences[idx+1:]...)
	}
}

// HorizontalPodAutoscalerSpec describes the desired functionality of the HorizontalPodAutoscaler.
type HorizontalPodAutoscalerSpec struct {
	// scaleTargetRef points to the target resource to scale, and is used to the pods for which metrics
	// should be collected, as well as to actually change the replica count.
	ScaleTargetRef CrossVersionObjectReference `json:"scaleTargetRef"`

	// minReplicas is the lower limit for the number of replicas to which the autoscaler
	// can scale down.  It defaults to 1 pod.  minReplicas is allowed to be 0 if the
	// alpha feature gate HPAScaleToZero is enabled and at least one Object or External
	// metric is configured.  Scaling is active as long as at least one metric value is
	// available.
	// +optional
	MinReplicas int32 `json:"minReplicas,omitempty"`

	// maxReplicas is the upper limit for the number of replicas to which the autoscaler can scale up.
	// It cannot be less that minReplicas.
	MaxReplicas int32 `json:"maxReplicas,omitempty"`

	// metrics contains the specifications for which to use to calculate the
	// desired replica count (the maximum replica count across all metrics will
	// be used).  The desired replica count is calculated multiplying the
	// ratio between the target value and the current value by the current
	// number of pods.  Ergo, metrics used must decrease as the pod count is
	// increased, and vice-versa.  See the individual metric source types for
	// more information about how each type of metric must respond.
	// +optional
	Metrics []MetricSpec `json:"metrics,omitempty"`

	// behavior configures the scaling behavior of the target
	// in both Up and Down directions (scaleUp and scaleDown fields respectively).
	// If not set, the default HPAScalingRules for scale up and scale down are used.
	// +optional
	Behavior *HorizontalPodAutoscalerBehavior `json:"behavior,omitempty"`
}

// HorizontalPodAutoscalerStatus describes the current status of a horizontal pod autoscaler.
type HorizontalPodAutoscalerStatus struct {
	// LastScaleTime is the last time the HorizontalPodAutoscaler scaled the number of pods,
	// used by the autoscaler to control how often the number of pods is changed.
	// +optional
	LastScaleTime types.Time `json:"lastScaleTime,omitempty"`

	// CurrentReplicas is current number of replicas of pods managed by this autoscaler,
	// as last seen by the autoscaler.
	CurrentReplicas int32 `json:"currentReplicas,omitempty"`

	// DesiredReplicas is the desired number of replicas of pods managed by this autoscaler,
	// as last calculated by the autoscaler.
	DesiredReplicas int32 `json:"desiredReplicas,omitempty"`
}

func (h *HorizontalPodAutoscalerStatus) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &h)
}

func (h *HorizontalPodAutoscalerStatus) JsonMarshal() ([]byte, error) {
	return json.Marshal(h)
}

// CrossVersionObjectReference contains enough information to let you identify the referred resource.
type CrossVersionObjectReference struct {
	// kind is the kind of the referent; More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds"
	Kind string `json:"kind"`

	// name is the name of the referent; More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	Name string `json:"name"`

	// apiVersion is the API version of the referent
	// +optional
	APIVersion string `json:"apiVersion,omitempty"`
}

// MetricSpec specifies how to scale based on a single metric
// (only `type` and one other matching field should be set at once).
type MetricSpec struct {
	// Type is the type of metric source.  It should be one of "Object",
	// "Pods" or "Resource", each mapping to a matching field in the object.
	Type MetricSourceType `json:"type"`

	//// Object refers to a metric describing a single kubernetes object
	//// (for example, hits-per-second on an Ingress object).
	//// +optional
	//Object *ObjectMetricSource `json:"object,omitempty"`
	//// Pods refers to a metric describing each pod in the current scale target
	//// (for example, transactions-processed-per-second).  The values will be
	//// averaged together before being compared to the target value.
	//// +optional
	//Pods *PodsMetricSource `json:"pods,omitempty"`
	// Resource refers to a resource metric (such as those specified in
	// requests and limits) known to Kubernetes describing each pod in the
	// current scale target (e.g. CPU or memory). Such metrics are built in to
	// Kubernetes, and have special scaling options on top of those available
	// to normal per-pod metrics using the "pods" source.
	// +optional
	Resource *ResourceMetricSource `json:"resource,omitempty"`
}

// MetricSourceType indicates the type of metric.
type MetricSourceType string

const (
	//// ObjectMetricSourceType is a metric describing a kubernetes object
	//// (for example, hits-per-second on an Ingress object).
	//ObjectMetricSourceType MetricSourceType = "Object"
	//// PodsMetricSourceType is a metric describing each pod in the current scale
	//// target (for example, transactions-processed-per-second).  The values
	//// will be averaged together before being compared to the target value.
	//PodsMetricSourceType MetricSourceType = "Pods"

	// ResourceMetricSourceType is a resource metric known to Kubernetes, as
	// specified in requests and limits, describing each pod in the current
	// scale target (e.g. CPU or memory).  Such metrics are built in to
	// Kubernetes, and have special scaling options on top of those available
	// to normal per-pod metrics (the "pods" source).
	ResourceMetricSourceType MetricSourceType = "Resource"
)

// ResourceMetricSource indicates how to scale on a resource metric known to
// Kubernetes, as specified in requests and limits, describing each pod in the
// current scale target (e.g. CPU or memory).  The values will be averaged
// together before being compared to the target.  Such metrics are built in to
// Kubernetes, and have special scaling options on top of those available to
// normal per-pod metrics using the "pods" source.  Only one "target" type
// should be set.
type ResourceMetricSource struct {
	// Name is the name of the resource in question.
	Name types.ResourceName `json:"name"`
	// Target specifies the target value for the given metric
	Target MetricTarget `json:"target"`
}

// MetricTarget defines the target value, average value, or average utilization of a specific metric
type MetricTarget struct {
	// Type represents whether the metric type is Utilization, Value, or AverageValue
	Type MetricTargetType `json:"type"`
	// Value is the target value of the metric (as a quantity).
	Value types.Quantity `json:"value,omitempty"`
	// TargetAverageValue is the target value of the average of the
	// metric across all relevant pods (as a quantity)
	AverageValue types.Quantity `json:"averageValue,omitempty"`

	// AverageUtilization is the target value of the average of the
	// resource metric across all relevant pods, represented as a percentage of
	// the requested value of the resource for the pods.
	// Currently only valid for Resource metric source type
	AverageUtilization int32 `json:"averageUtilization,omitempty"`
}

// MetricTargetType specifies the type of metric being targeted, and should be either
// "Value", "AverageValue", or "Utilization"
type MetricTargetType string

const (
	// UtilizationMetricType is a possible value for MetricTarget.Type.
	UtilizationMetricType MetricTargetType = "Utilization"
	// ValueMetricType is a possible value for MetricTarget.Type.
	ValueMetricType MetricTargetType = "Value"
	// AverageValueMetricType is a possible value for MetricTarget.Type.
	AverageValueMetricType MetricTargetType = "AverageValue"
)

// HorizontalPodAutoscalerBehavior configures a scaling behavior for Up and Down direction
// (scaleUp and scaleDown fields respectively).
type HorizontalPodAutoscalerBehavior struct {
	// scaleUp is scaling policy for scaling Up.
	// If not set, the default value is the higher of:
	//   * increase no more than 1 pod per 15 seconds
	//   * double the number of pods per 60 seconds
	// No stabilization is used.
	// +optional
	ScaleUp *HPAScalingRules `json:"scaleUp,omitempty"`
	// scaleDown is scaling policy for scaling Down.
	// If not set, the default value is to allow to scale down to minReplicas pods, with a
	// 300 second stabilization window (i.e., the highest recommendation for
	// the last 300sec is used).
	// +optional
	ScaleDown *HPAScalingRules `json:"scaleDown,omitempty"`
}

func DefaultScaleUpRule() (r *HPAScalingRules) {
	upPolicies := make([]HPAScalingPolicy, 2)
	upPolicies[0] = HPAScalingPolicy{
		Type:          PercentScalingPolicy,
		Value:         100,
		PeriodSeconds: 60,
	}
	upPolicies[1] = HPAScalingPolicy{
		Type:          PodsScalingPolicy,
		Value:         1,
		PeriodSeconds: 15,
	}
	return &HPAScalingRules{
		StabilizationWindowSeconds: 0,
		SelectPolicy:               MaxPolicySelect,
		Policies:                   upPolicies,
	}
}

func DefaultScaleDownRule() (r *HPAScalingRules) {
	downPolicies := make([]HPAScalingPolicy, 1)
	downPolicies[0] = HPAScalingPolicy{
		Type:          PercentScalingPolicy,
		Value:         100,
		PeriodSeconds: 15,
	}
	return &HPAScalingRules{
		StabilizationWindowSeconds: 300,
		SelectPolicy:               MinPolicySelect,
		Policies:                   downPolicies,
	}
}

// ScalingPolicySelect is used to specify which policy should be used while scaling in a certain direction
type ScalingPolicySelect string

const (
	// MaxPolicySelect selects the policy with the highest possible change.
	MaxPolicySelect ScalingPolicySelect = "Max"
	// MinPolicySelect selects the policy with the lowest possible change.
	MinPolicySelect ScalingPolicySelect = "Min"
	// DisabledPolicySelect disables the scaling in this direction.
	DisabledPolicySelect ScalingPolicySelect = "Disabled"
)

// HPAScalingRules configures the scaling behavior for one direction.
// These Rules are applied after calculating DesiredReplicas from metrics for the HPA.
// They can limit the scaling velocity by specifying scaling policies.
// They can prevent flapping by specifying the stabilization window, so that the
// number of replicas is not set instantly, instead, the safest value from the stabilization
// window is chosen.
type HPAScalingRules struct {
	// StabilizationWindowSeconds is the number of seconds for which past recommendations should be
	// considered while scaling up or scaling down.
	// StabilizationWindowSeconds must be greater than or equal to zero and less than or equal to 3600 (one hour).
	// If not set, use the default values:
	// - For scale up: 0 (i.e. no stabilization is done).
	// - For scale down: 300 (i.e. the stabilization window is 300 seconds long).
	// +optional
	StabilizationWindowSeconds int32 `json:"stabilizationWindowSeconds,omitempty"`
	// selectPolicy is used to specify which policy should be used.
	// If not set, the default value MaxPolicySelect is used.
	// +optional
	SelectPolicy ScalingPolicySelect `json:"selectPolicy,omitempty"`
	// policies is a list of potential scaling polices which can used during scaling.
	// At least one policy must be specified, otherwise the HPAScalingRules will be discarded as invalid
	// +optional
	Policies []HPAScalingPolicy `json:"policies,omitempty"`
}

// HPAScalingPolicyType is the type of the policy which could be used while making scaling decisions.
type HPAScalingPolicyType string

const (
	// PodsScalingPolicy is a policy used to specify a change in absolute number of pods.
	PodsScalingPolicy HPAScalingPolicyType = "Pods"
	// PercentScalingPolicy is a policy used to specify a relative amount of change with respect to
	// the current number of pods.
	PercentScalingPolicy HPAScalingPolicyType = "Percent"
)

// HPAScalingPolicy is a single policy which must hold true for a specified past interval.
type HPAScalingPolicy struct {
	// Type is used to specify the scaling policy.
	Type HPAScalingPolicyType `json:"type"`
	// Value contains the amount of change which is permitted by the policy.
	// It must be greater than zero
	Value int32 `json:"value"`
	// PeriodSeconds specifies the window of time for which the policy should hold true.
	// PeriodSeconds must be greater than zero and less than or equal to 1800 (30 min).
	PeriodSeconds int32 `json:"periodSeconds"`
}

// HorizontalPodAutoscalerList is a list of HorizontalPodAutoscalers.
type HorizontalPodAutoscalerList struct {
	meta.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	meta.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// List of pods.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md
	Items []HorizontalPodAutoscaler `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func (h *HorizontalPodAutoscalerList) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &h)
}

func (h *HorizontalPodAutoscalerList) JsonMarshal() ([]byte, error) {
	return json.Marshal(h)
}

func (h *HorizontalPodAutoscalerList) AddItemFromStr(objectStr string) error {
	object := &HorizontalPodAutoscaler{}
	buf, err := strconv.Unquote(objectStr)
	err = object.JsonUnmarshal([]byte(buf))
	if err != nil {
		return err
	}
	h.Items = append(h.Items, *object)
	return nil
}

func (h *HorizontalPodAutoscalerList) AppendItemsFromStr(objectStrs []string) error {
	for _, obj := range objectStrs {
		object := &HorizontalPodAutoscaler{}
		err := object.JsonUnmarshal([]byte(obj))
		if err != nil {
			return err
		}
		h.Items = append(h.Items, *object)
	}
	return nil
}

func (h *HorizontalPodAutoscalerList) GetItems() any {
	return h.Items
}

func (h *HorizontalPodAutoscalerList) GetIApiObjectArr() (res []IApiObject) {
	for _, item := range h.Items {
		itemTemp := item
		res = append(res, &itemTemp)
	}
	return res
}

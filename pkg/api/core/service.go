package core

import (
	"encoding/json"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/types"
	"strconv"
)

// Service is a named abstraction of software service (for example, mysql) consisting of local port
// (for example 3306) that the proxy listens on, and the selector that determines which pods
// will answer requests sent through the proxy.
type Service struct {
	meta.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	meta.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines the behavior of a service.
	// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec ServiceSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Most recently observed status of the service.
	// Populated by the system.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Status ServiceStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

func (s *Service) DeleteOwnerReference(uid types.UID) {
	has := false
	idx := 0
	for i, o := range s.OwnerReferences {
		if o.UID == uid {
			has = true
			idx = i
			break
		}
	}
	if has {
		s.OwnerReferences = append(s.OwnerReferences[:idx], s.OwnerReferences[idx+1:]...)
	}
}

func (s *Service) AppendOwnerReference(reference meta.OwnerReference) {
	s.OwnerReferences = append(s.OwnerReferences, reference)
}

func (s *Service) GenerateOwnerReference() meta.OwnerReference {
	return meta.OwnerReference{
		APIVersion: s.APIVersion,
		Kind:       s.Kind,
		Name:       s.Name,
		UID:        s.UID,
		Controller: false,
	}
}

func (s *Service) CreateFromEtcdString(str string) error {
	return s.JsonUnmarshal([]byte(str))
}

func (s *Service) SetResourceVersion(version string) {
	s.ObjectMeta.ResourceVersion = version
}

func (s *Service) GetResourceVersion() string {
	return s.ObjectMeta.ResourceVersion
}

func (s *Service) SetStatus(st IApiObjectStatus) bool {
	status, ok := st.(*ServiceStatus)
	if ok {
		s.Status = *status
	}
	return ok
}

func (s *Service) GetStatus() IApiObjectStatus {
	return &s.Status
}

func (s *Service) JsonUnmarshalStatus(data []byte) error {
	return json.Unmarshal(data, &(s.Status))
}

func (s *Service) JsonMarshalStatus() ([]byte, error) {
	return json.Marshal(s.Status)
}

func (s *Service) JsonMarshal() ([]byte, error) {
	return json.Marshal(s)
}

func (s *Service) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &s)
}

func (s *Service) SetUID(uid types.UID) {
	s.ObjectMeta.UID = uid
}

func (s *Service) GetUID() types.UID {
	return s.ObjectMeta.UID
}

// ServiceSpec describes the attributes that a user creates on a service.
type ServiceSpec struct {
	// The list of ports that are exposed by this service.
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies
	// +patchMergeKey=port
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=port
	// +listMapKey=protocol
	Ports []ServicePort `json:"ports,omitempty" patchStrategy:"merge" patchMergeKey:"port" protobuf:"bytes,1,rep,name=ports"`

	// Route service traffic to pods with label keys and values matching this
	// selector. If empty or not present, the service is assumed to have an
	// external process managing its endpoints, which Kubernetes will not
	// modify. Only applies to types ClusterIP, NodePort, and LoadBalancer.
	// Ignored if type is ExternalName.
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/
	// +optional
	// +mapType=atomic
	Selector map[string]string `json:"selector,omitempty" protobuf:"bytes,2,rep,name=selector"`

	// clusterIP is the IP address of the service and is usually assigned
	// randomly. If an address is specified manually, is in-range (as per
	// system configuration), and is not in use, it will be allocated to the
	// service; otherwise creation of the service will fail. This field may not
	// be changed through updates unless the type field is also being changed
	// to ExternalName (which requires this field to be blank) or the type
	// field is being changed from ExternalName (in which case this field may
	// optionally be specified, as describe above).  Valid values are "None",
	// empty string (""), or a valid IP address. Setting this to "None" makes a
	// "headless service" (no virtual IP), which is useful when direct endpoint
	// connections are preferred and proxying is not required.  Only applies to
	// types ClusterIP, NodePort, and LoadBalancer. If this field is specified
	// when creating a Service of type ExternalName, creation will fail. This
	// field will be wiped when updating a Service to type ExternalName.
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies
	// +optional
	ClusterIP string `json:"clusterIP,omitempty" protobuf:"bytes,3,opt,name=clusterIP"`

	// type determines how the Service is exposed. Defaults to ClusterIP. Valid
	// options are ExternalName, ClusterIP, NodePort, and LoadBalancer.
	// "ClusterIP" allocates a cluster-internal IP address for load-balancing
	// to endpoints. Endpoints are determined by the selector or if that is not
	// specified, by manual construction of an Endpoints object or
	// EndpointSlice objects. If clusterIP is "None", no virtual IP is
	// allocated and the endpoints are published as a set of endpoints rather
	// than a virtual IP.
	// "NodePort" builds on ClusterIP and allocates a port on every node which
	// routes to the same endpoints as the clusterIP.
	// "LoadBalancer" builds on NodePort and creates an external load-balancer
	// (if supported in the current cloud) which routes to the same endpoints
	// as the clusterIP.
	// "ExternalName" aliases this service to the specified externalName.
	// Several other fields do not apply to ExternalName services.
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#publishing-services-service-types
	// +optional
	Type ServiceType `json:"type,omitempty" protobuf:"bytes,4,opt,name=type,casttype=ServiceType"`
}

// Service Type string describes ingress methods for a service
// +enum
type ServiceType string

const (
	// ServiceTypeClusterIP means a service will only be accessible inside the
	// cluster, via the cluster IP.
	ServiceTypeClusterIP ServiceType = "ClusterIP"

	// ServiceTypeNodePort means a service will be exposed on one port of
	// every node, in addition to 'ClusterIP' type.
	//ServiceTypeNodePort ServiceType = "NodePort"

	// ServiceTypeLoadBalancer means a service will be exposed via an
	// external load balancer (if the cloud provider supports it), in addition
	// to 'NodePort' type.
	//ServiceTypeLoadBalancer ServiceType = "LoadBalancer"

	// ServiceTypeExternalName means a service consists of only a reference to
	// an external name that kubedns or equivalent will return as a CNAME
	// record, with no exposing or proxying of any pods involved.
	//ServiceTypeExternalName ServiceType = "ExternalName"
)

// ServicePort contains information on service's port.
type ServicePort struct {
	// The name of this port within the service. This must be a DNS_LABEL.
	// All ports within a ServiceSpec must have unique names. When considering
	// the endpoints for a Service, this must match the 'name' field in the
	// EndpointPort.
	// Optional if only one ServicePort is defined on this service.
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`

	// The IP protocol for this port. Supports "TCP", "UDP", and "SCTP".
	// Default is TCP.
	// +default="TCP"
	// +optional
	//Protocol Protocol `json:"protocol,omitempty" protobuf:"bytes,2,opt,name=protocol,casttype=Protocol"`

	// The port that will be exposed by this service.
	Port int32 `json:"port" protobuf:"varint,3,opt,name=port"`

	// Number or name of the port to access on the pods targeted by the service.
	// Number must be in the range 1 to 65535. Name must be an IANA_SVC_NAME.
	// If this is a string, it will be looked up as a named port in the
	// target Pod's container ports. If this is not specified, the value
	// of the 'port' field is used (an identity map).
	// This field is ignored for services with clusterIP=None, and should be
	// omitted or set equal to the 'port' field.
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service
	// +optional
	TargetPort int32 `json:"targetPort,omitempty" protobuf:"bytes,4,opt,name=targetPort"`

	// The port on each node on which this service is exposed when type is
	// NodePort or LoadBalancer.  Usually assigned by the system. If a value is
	// specified, in-range, and not in use it will be used, otherwise the
	// operation will fail.  If not specified, a port will be allocated if this
	// Service requires one.  If this field is specified when creating a
	// Service which does not need it, creation will fail. This field will be
	// wiped when updating a Service to no longer need it (e.g. changing type
	// from NodePort to ClusterIP).
	// More info: https://kubernetes.io/docs/concepts/services-networking/service/#type-nodeport
	// +optional
	//NodePort int32 `json:"nodePort,omitempty" protobuf:"varint,5,opt,name=nodePort"`
}

// ServiceStatus represents the current status of a service.
type ServiceStatus struct {
}

func (s *ServiceStatus) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &s)
}

func (s *ServiceStatus) JsonMarshal() ([]byte, error) {
	return json.Marshal(s)
}

// ServiceList holds a list of services.
type ServiceList struct {
	meta.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	meta.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// List of services
	Items []Service `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func (s *ServiceList) GetIApiObjectArr() (res []IApiObject) {
	for _, item := range s.Items {
		itemTemp := item
		res = append(res, &itemTemp)
	}
	return res
}

func (s *ServiceList) AppendItemsFromStr(objectStrs []string) error {
	for _, obj := range objectStrs {
		object := &Service{}
		err := object.JsonUnmarshal([]byte(obj))
		if err != nil {
			return err
		}
		s.Items = append(s.Items, *object)
	}
	return nil
}

func (s *ServiceList) AddItemFromStr(objectStr string) error {
	object := &Service{}
	buf, err := strconv.Unquote(objectStr)
	err = object.JsonUnmarshal([]byte(buf))
	if err != nil {
		return err
	}
	s.Items = append(s.Items, *object)
	return nil
}

func (s *ServiceList) GetItems() any {
	return s.Items
}

func (s *ServiceList) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &s)
}

func (s *ServiceList) JsonMarshal() ([]byte, error) {
	return json.Marshal(s)
}

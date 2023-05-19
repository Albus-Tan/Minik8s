package core

import (
	"encoding/json"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/types"
	"strconv"
)

// Node is a worker node in Kubernetes.
// Each node will have a unique identifier in the cache (i.e. in etcd).
type Node struct {
	meta.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	meta.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines the behavior of a node.
	// https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec NodeSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Most recently observed status of the node.
	// Populated by the system.
	// Read-only.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Status NodeStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

func (n *Node) DeleteOwnerReference(uid types.UID) {
	has := false
	idx := 0
	for i, o := range n.OwnerReferences {
		if o.UID == uid {
			has = true
			idx = i
			break
		}
	}
	if has {
		n.OwnerReferences = append(n.OwnerReferences[:idx], n.OwnerReferences[idx+1:]...)
	}
}

func (n *Node) AppendOwnerReference(reference meta.OwnerReference) {
	n.OwnerReferences = append(n.OwnerReferences, reference)
}

func (n *Node) GenerateOwnerReference() meta.OwnerReference {
	return meta.OwnerReference{
		APIVersion: n.APIVersion,
		Kind:       n.Kind,
		Name:       n.Name,
		UID:        n.UID,
		Controller: false,
	}
}

func (n *Node) CreateFromEtcdString(str string) error {
	return n.JsonUnmarshal([]byte(str))
}

func (n *Node) SetResourceVersion(version string) {
	n.ObjectMeta.ResourceVersion = version
}

func (n *Node) GetResourceVersion() string {
	return n.ObjectMeta.ResourceVersion
}

func (n *Node) SetStatus(s IApiObjectStatus) bool {
	status, ok := s.(*NodeStatus)
	if ok {
		n.Status = *status
	}
	return ok
}

func (n *Node) GetStatus() IApiObjectStatus {
	return &n.Status
}

func (n *Node) JsonUnmarshalStatus(data []byte) error {
	return json.Unmarshal(data, &(n.Status))
}

func (n *Node) JsonMarshalStatus() ([]byte, error) {
	return json.Marshal(n.Status)
}

func (n *Node) JsonMarshal() ([]byte, error) {
	return json.Marshal(n)
}

func (n *Node) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &n)
}

func (n *Node) SetUID(uid types.UID) {
	n.ObjectMeta.UID = uid
}

func (n *Node) GetUID() types.UID {
	return n.ObjectMeta.UID
}

// NodeSpec describes the attributes that a node is created with.
type NodeSpec struct {
	// PodCIDR represents the pod IP range assigned to the node.
	// +optional
	PodCIDR string `json:"podCIDR,omitempty" protobuf:"bytes,1,opt,name=podCIDR"`

	// podCIDRs represents the IP ranges assigned to the node for usage by Pods on that node. If this
	// field is specified, the 0th entry must match the podCIDR field. It may contain at most 1 value for
	// each of IPv4 and IPv6.
	// +optional
	// +patchStrategy=merge
	PodCIDRs []string `json:"podCIDRs,omitempty" protobuf:"bytes,7,opt,name=podCIDRs" patchStrategy:"merge"`

	// Address represents the node IP address
	Address string `json:"address,omitempty"`
}

// NodeStatus is information about the current status of a node.
type NodeStatus struct {
	// NodePhase is the recently observed lifecycle phase of the node.
	// More info: https://kubernetes.io/docs/concepts/nodes/node/#phase
	// The field is never populated, and now is deprecated.
	// +optional
	Phase NodePhase `json:"phase,omitempty" protobuf:"bytes,3,opt,name=phase,casttype=NodePhase"`
	// List of addresses reachable to the node.
	// Queried from cloud provider, if available.
	// More info: https://kubernetes.io/docs/concepts/nodes/node/#addresses
	// Note: This field is declared as mergeable, but the merge key is not sufficiently
	// unique, which can cause data corruption when it is merged. Callers should instead
	// use a full-replacement patch. See https://pr.k8s.io/79391 for an example.
	// Consumers should assume that addresses can change during the
	// lifetime of a Node. However, there are some exceptions where this may not
	// be possible, such as Pods that inherit a Node's address in its own status or
	// consumers of the downward API (status.hostIP).
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Addresses []NodeAddress `json:"addresses,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,5,rep,name=addresses"`
}

func (n *NodeStatus) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &n)
}

func (n *NodeStatus) JsonMarshal() ([]byte, error) {
	return json.Marshal(n)
}

// +enum

type NodePhase string

// These are the valid phases of node.
const (
	// NodePending means the node has been created/added by the system, but not configured.
	NodePending NodePhase = "Pending"
	// NodeRunning means the node has been configured and has Kubernetes components running.
	NodeRunning NodePhase = "Running"
	// NodeTerminated means the node has been removed from the cluster.
	NodeTerminated NodePhase = "Terminated"
)

// NodeAddress contains information for the node's address.
type NodeAddress struct {
	// Node address type, one of Hostname, ExternalIP or InternalIP.
	Type NodeAddressType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=NodeAddressType"`
	// The node address.
	Address string `json:"address" protobuf:"bytes,2,opt,name=address"`
}

type NodeAddressType string

// These are built-in addresses type of node. A cloud provider may set a type not listed here.
const (
	// NodeHostName identifies a name of the node. Although every node can be assumed
	// to have a NodeAddress of this type, its exact syntax and semantics are not
	// defined, and are not consistent between different clusters.
	NodeHostName NodeAddressType = "Hostname"

	// NodeInternalIP identifies an IP address which is assigned to one of the node's
	// network interfaces. Every node should have at least one address of this type.
	//
	// An internal IP is normally expected to be reachable from every other node, but
	// may not be visible to hosts outside the cluster. By default it is assumed that
	// kube-apiserver can reach node internal IPs, though it is possible to configure
	// clusters where this is not the case.
	//
	// NodeInternalIP is the default type of node IP, and does not necessarily imply
	// that the IP is ONLY reachable internally. If a node has multiple internal IPs,
	// no specific semantics are assigned to the additional IPs.
	NodeInternalIP NodeAddressType = "InternalIP"

	// NodeExternalIP identifies an IP address which is, in some way, intended to be
	// more usable from outside the cluster then an internal IP, though no specific
	// semantics are defined. It may be a globally routable IP, though it is not
	// required to be.
	//
	// External IPs may be assigned directly to an interface on the node, like a
	// NodeInternalIP, or alternatively, packets sent to the external IP may be NAT'ed
	// to an internal node IP rather than being delivered directly (making the IP less
	// efficient for node-to-node traffic than a NodeInternalIP).
	NodeExternalIP NodeAddressType = "ExternalIP"

	// NodeInternalDNS identifies a DNS name which resolves to an IP address which has
	// the characteristics of a NodeInternalIP. The IP it resolves to may or may not
	// be a listed NodeInternalIP address.
	NodeInternalDNS NodeAddressType = "InternalDNS"

	// NodeExternalDNS identifies a DNS name which resolves to an IP address which has
	// the characteristics of a NodeExternalIP. The IP it resolves to may or may not
	// be a listed NodeExternalIP address.
	NodeExternalDNS NodeAddressType = "ExternalDNS"
)

// NodeList is the whole list of all Nodes which have been registered with master.
type NodeList struct {
	meta.TypeMeta `json:",inline"`
	// Standard list metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	// +optional
	meta.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// List of nodes
	Items []Node `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func (n *NodeList) GetIApiObjectArr() (res []IApiObject) {
	for _, item := range n.Items {
		itemTemp := item
		res = append(res, &itemTemp)
	}
	return res
}

func (n *NodeList) AppendItemsFromStr(objectStrs []string) error {
	for _, obj := range objectStrs {
		object := &Node{}
		err := object.JsonUnmarshal([]byte(obj))
		if err != nil {
			return err
		}
		n.Items = append(n.Items, *object)
	}
	return nil
}

func (n *NodeList) AddItemFromStr(objectStr string) error {
	object := &Node{}
	buf, err := strconv.Unquote(objectStr)
	err = object.JsonUnmarshal([]byte(buf))
	if err != nil {
		return err
	}
	n.Items = append(n.Items, *object)
	return nil
}

func (n *NodeList) GetItems() any {
	return n.Items
}

func (n *NodeList) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &n)
}

func (n *NodeList) JsonMarshal() ([]byte, error) {
	return json.Marshal(n)
}

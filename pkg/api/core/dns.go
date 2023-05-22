package core

import (
	"encoding/json"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/types"
	"strconv"
)

type DNS struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec            DnsSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status          DnsStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

func (j *DNS) SetUID(uid types.UID) {
	j.ObjectMeta.UID = uid
}

func (j *DNS) GetUID() types.UID {
	return j.ObjectMeta.UID
}

func (j *DNS) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &j)
}

func (j *DNS) JsonMarshal() ([]byte, error) {
	return json.Marshal(j)
}

func (j *DNS) JsonUnmarshalStatus(data []byte) error {
	return json.Unmarshal(data, &(j.Status))
}

func (j *DNS) JsonMarshalStatus() ([]byte, error) {
	return json.Marshal(j.Status)
}

func (j *DNS) SetStatus(s IApiObjectStatus) bool {
	status, ok := s.(*DnsStatus)
	if ok {
		j.Status = *status
	}
	return ok
}

func (j *DNS) GetStatus() IApiObjectStatus {
	return &j.Status
}

func (j *DNS) GetResourceVersion() string {
	return j.ObjectMeta.ResourceVersion
}

func (j *DNS) SetResourceVersion(version string) {
	j.ObjectMeta.ResourceVersion = version
}

func (j *DNS) CreateFromEtcdString(str string) error {
	return j.JsonUnmarshal([]byte(str))
}

func (j *DNS) GenerateOwnerReference() meta.OwnerReference {
	return meta.OwnerReference{
		APIVersion: j.APIVersion,
		Kind:       j.Kind,
		Name:       j.Name,
		UID:        j.UID,
		Controller: false,
	}
}

func (j *DNS) AppendOwnerReference(reference meta.OwnerReference) {
	j.OwnerReferences = append(j.OwnerReferences, reference)
}

func (j *DNS) DeleteOwnerReference(uid types.UID) {
	has := false
	idx := 0
	for i, o := range j.OwnerReferences {
		if o.UID == uid {
			has = true
			idx = i
			break
		}
	}
	if has {
		j.OwnerReferences = append(j.OwnerReferences[:idx], j.OwnerReferences[idx+1:]...)
	}
}

type DnsSpec struct {
	Hostname string       `json:"hostname,omitempty"`
	Mappings []DnsMapping `json:"mappings,omitempty"`
}

type DnsMapping struct {
	Address string `json:"address,omitempty"`
	Path    string `json:"path,omitempty"`
}

type DnsStatus struct {
}

func (j *DnsStatus) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &j)
}

func (j *DnsStatus) JsonMarshal() ([]byte, error) {
	return json.Marshal(j)
}

type DnsList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items         []DNS `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func (j *DnsList) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &j)
}

func (j *DnsList) JsonMarshal() ([]byte, error) {
	return json.Marshal(j)
}

func (j *DnsList) AddItemFromStr(objectStr string) error {
	object := &DNS{}
	buf, err := strconv.Unquote(objectStr)
	err = object.JsonUnmarshal([]byte(buf))
	if err != nil {
		return err
	}
	j.Items = append(j.Items, *object)
	return nil
}

func (j *DnsList) AppendItemsFromStr(objectStrs []string) error {
	for _, obj := range objectStrs {
		object := &DNS{}
		err := object.JsonUnmarshal([]byte(obj))
		if err != nil {
			return err
		}
		j.Items = append(j.Items, *object)
	}
	return nil
}

func (j *DnsList) GetItems() any {
	return j.Items
}

func (j *DnsList) GetIApiObjectArr() (res []IApiObject) {
	for _, item := range j.Items {
		itemTemp := item
		res = append(res, &itemTemp)
	}
	return res
}

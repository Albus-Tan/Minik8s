package core

import (
	"encoding/json"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/types"
	"strconv"
)

type Heartbeat struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec            HeartbeatSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status          HeartbeatStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

func (j *Heartbeat) PrintBrief() {
	//TODO implement me
	panic("implement me")
}

func (j *Heartbeat) SetUID(uid types.UID) {
	j.ObjectMeta.UID = uid
}

func (j *Heartbeat) GetUID() types.UID {
	return j.ObjectMeta.UID
}

func (j *Heartbeat) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &j)
}

func (j *Heartbeat) JsonMarshal() ([]byte, error) {
	return json.Marshal(j)
}

func (j *Heartbeat) JsonUnmarshalStatus(data []byte) error {
	return json.Unmarshal(data, &(j.Status))
}

func (j *Heartbeat) JsonMarshalStatus() ([]byte, error) {
	return json.Marshal(j.Status)
}

func (j *Heartbeat) SetStatus(s IApiObjectStatus) bool {
	status, ok := s.(*HeartbeatStatus)
	if ok {
		j.Status = *status
	}
	return ok
}

func (j *Heartbeat) GetStatus() IApiObjectStatus {
	return &j.Status
}

func (j *Heartbeat) GetResourceVersion() string {
	return j.ObjectMeta.ResourceVersion
}

func (j *Heartbeat) SetResourceVersion(version string) {
	j.ObjectMeta.ResourceVersion = version
}

func (j *Heartbeat) CreateFromEtcdString(str string) error {
	return j.JsonUnmarshal([]byte(str))
}

func (j *Heartbeat) GenerateOwnerReference() meta.OwnerReference {
	return meta.OwnerReference{
		APIVersion: j.APIVersion,
		Kind:       j.Kind,
		Name:       j.Name,
		UID:        j.UID,
		Controller: false,
	}
}

func (j *Heartbeat) AppendOwnerReference(reference meta.OwnerReference) {
	j.OwnerReferences = append(j.OwnerReferences, reference)
}

func (j *Heartbeat) DeleteOwnerReference(uid types.UID) {
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

type HeartbeatSpec struct {
	NodeUID string `json:"nodeUID,omitempty"`
}

type HeartbeatStatus struct {
	HeartbeatID string     `json:"heartbeatID,omitempty"`
	Timestamp   types.Time `json:"timestamp,omitempty"`
}

func (j *HeartbeatStatus) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &j)
}

func (j *HeartbeatStatus) JsonMarshal() ([]byte, error) {
	return json.Marshal(j)
}

type HeartbeatList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items         []Heartbeat `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func (j *HeartbeatList) PrintBrief() {
	//TODO implement me
	panic("implement me")
}

func (j *HeartbeatList) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &j)
}

func (j *HeartbeatList) JsonMarshal() ([]byte, error) {
	return json.Marshal(j)
}

func (j *HeartbeatList) AddItemFromStr(objectStr string) error {
	object := &Heartbeat{}
	buf, err := strconv.Unquote(objectStr)
	err = object.JsonUnmarshal([]byte(buf))
	if err != nil {
		return err
	}
	j.Items = append(j.Items, *object)
	return nil
}

func (j *HeartbeatList) AppendItemsFromStr(objectStrs []string) error {
	for _, obj := range objectStrs {
		object := &Heartbeat{}
		err := object.JsonUnmarshal([]byte(obj))
		if err != nil {
			return err
		}
		j.Items = append(j.Items, *object)
	}
	return nil
}

func (j *HeartbeatList) GetItems() any {
	return j.Items
}

func (j *HeartbeatList) GetIApiObjectArr() (res []IApiObject) {
	for _, item := range j.Items {
		itemTemp := item
		res = append(res, &itemTemp)
	}
	return res
}

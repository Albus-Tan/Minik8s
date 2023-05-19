package core

import (
	"encoding/json"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/types"
	"strconv"
)

type Func struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec            FuncSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status          FuncStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

func (f *Func) SetUID(uid types.UID) {
	f.ObjectMeta.UID = uid
}

func (f *Func) GetUID() types.UID {
	return f.ObjectMeta.UID
}

func (f *Func) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &f)
}

func (f *Func) JsonMarshal() ([]byte, error) {
	return json.Marshal(f)
}

func (f *Func) JsonUnmarshalStatus(data []byte) error {
	return json.Unmarshal(data, &(f.Status))
}

func (f *Func) JsonMarshalStatus() ([]byte, error) {
	return json.Marshal(f.Status)
}

func (f *Func) SetStatus(s IApiObjectStatus) bool {
	status, ok := s.(*FuncStatus)
	if ok {
		f.Status = *status
	}
	return ok
}

func (f *Func) GetStatus() IApiObjectStatus {
	return &f.Status
}

func (f *Func) GetResourceVersion() string {
	return f.ObjectMeta.ResourceVersion
}

func (f *Func) SetResourceVersion(version string) {
	f.ObjectMeta.ResourceVersion = version
}

func (f *Func) CreateFromEtcdString(str string) error {
	return f.JsonUnmarshal([]byte(str))
}

func (f *Func) GenerateOwnerReference() meta.OwnerReference {
	return meta.OwnerReference{
		APIVersion: f.APIVersion,
		Kind:       f.Kind,
		Name:       f.Name,
		UID:        f.UID,
		Controller: false,
	}
}

func (f *Func) AppendOwnerReference(reference meta.OwnerReference) {
	f.OwnerReferences = append(f.OwnerReferences, reference)
}

func (f *Func) DeleteOwnerReference(uid types.UID) {
	has := false
	idx := 0
	for i, o := range f.OwnerReferences {
		if o.UID == uid {
			has = true
			idx = i
			break
		}
	}
	if has {
		f.OwnerReferences = append(f.OwnerReferences[:idx], f.OwnerReferences[idx+1:]...)
	}
}

type FuncSpec struct {
	// Name is unique for Func, should be same as Name in ObjectMeta field
	Name string `json:"name"`

	PreRun   string `json:"preRun"`
	Function string `json:"function"`
	Left     string `json:"left"`
	Right    string `json:"right"`
}

type FuncStatus struct {
	InstanceId string `json:"instanceId,omitempty"`
}

func (f *FuncStatus) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &f)
}

func (f *FuncStatus) JsonMarshal() ([]byte, error) {
	return json.Marshal(f)
}

type FuncList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items         []Func `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func (f *FuncList) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &f)
}

func (f *FuncList) JsonMarshal() ([]byte, error) {
	return json.Marshal(f)
}

func (f *FuncList) AddItemFromStr(objectStr string) error {
	object := &Func{}
	buf, err := strconv.Unquote(objectStr)
	err = object.JsonUnmarshal([]byte(buf))
	if err != nil {
		return err
	}
	f.Items = append(f.Items, *object)
	return nil
}

func (f *FuncList) AppendItemsFromStr(objectStrs []string) error {
	for _, obj := range objectStrs {
		object := &Func{}
		err := object.JsonUnmarshal([]byte(obj))
		if err != nil {
			return err
		}
		f.Items = append(f.Items, *object)
	}
	return nil
}

func (f *FuncList) GetItems() any {
	return f.Items
}

func (f *FuncList) GetIApiObjectArr() []IApiObject {
	//TODO implement me
	panic("implement me")
}

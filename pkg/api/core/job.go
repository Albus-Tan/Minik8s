package core

import (
	"encoding/json"
	"minik8s/pkg/api/meta"
	"minik8s/pkg/api/types"
	"strconv"
)

type Job struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Spec            JobSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status          JobStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

func (j *Job) SetUID(uid types.UID) {
	j.ObjectMeta.UID = uid
}

func (j *Job) GetUID() types.UID {
	return j.ObjectMeta.UID
}

func (j *Job) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &j)
}

func (j *Job) JsonMarshal() ([]byte, error) {
	return json.Marshal(j)
}

func (j *Job) JsonUnmarshalStatus(data []byte) error {
	return json.Unmarshal(data, &(j.Status))
}

func (j *Job) JsonMarshalStatus() ([]byte, error) {
	return json.Marshal(j.Status)
}

func (j *Job) SetStatus(s IApiObjectStatus) bool {
	status, ok := s.(*JobStatus)
	if ok {
		j.Status = *status
	}
	return ok
}

func (j *Job) GetStatus() IApiObjectStatus {
	return &j.Status
}

func (j *Job) GetResourceVersion() string {
	return j.ObjectMeta.ResourceVersion
}

func (j *Job) SetResourceVersion(version string) {
	j.ObjectMeta.ResourceVersion = version
}

func (j *Job) CreateFromEtcdString(str string) error {
	return j.JsonUnmarshal([]byte(str))
}

func (j *Job) GenerateOwnerReference() meta.OwnerReference {
	return meta.OwnerReference{
		APIVersion: j.APIVersion,
		Kind:       j.Kind,
		Name:       j.Name,
		UID:        j.UID,
		Controller: false,
	}
}

func (j *Job) AppendOwnerReference(reference meta.OwnerReference) {
	j.OwnerReferences = append(j.OwnerReferences, reference)
}

func (j *Job) DeleteOwnerReference(uid types.UID) {
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

type JobSpec struct {
	CuFilePath     string  `json:"cuFilePath,omitempty"`
	ResultFileName string  `json:"resultFileName,omitempty"`
	ResultFilePath string  `json:"resultFilePath,omitempty"`
	Args           JobArgs `json:"args,omitempty"`
}

type JobArgs struct {
	NumTasksPerNode int         `json:"numTasksPerNode,omitempty"`
	CpusPerTask     int         `json:"cpusPerTask,omitempty"`
	GpuResources    int         `json:"gpuResources,omitempty"`
	Mail            *MailRemind `json:"mail,omitempty"`
}

type MailRemind struct {
	Type     MailRemindType `json:"type,omitempty"`
	UserName string         `json:"userName,omitempty"`
}

type MailRemindType string

const (
	MailRemindAll   MailRemindType = "all"
	MailRemindBegin MailRemindType = "begin"
	MailRemindEnd   MailRemindType = "end"
	MailRemindFail  MailRemindType = "fail"
)

type JobStatus struct {
	JobID string   `json:"jobID,omitempty"`
	State JobState `json:"state,omitempty"`
}

func (j *JobStatus) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &j)
}

func (j *JobStatus) JsonMarshal() ([]byte, error) {
	return json.Marshal(j)
}

type JobState string

const (
	JobPending   JobState = "PENDING"
	JobRunning   JobState = "RUNNING"
	JobFailed    JobState = "FAILED"
	JobCompleted JobState = "COMPLETED"
	JobMissing   JobState = "MISSING"
)

func JobUnfinished(state JobState) bool {
	return state == JobPending || state == JobRunning
}

func JobFinished(state JobState) bool {
	return state == JobFailed || state == JobCompleted
}

type JobList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items         []Job `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func (j *JobList) JsonUnmarshal(data []byte) error {
	return json.Unmarshal(data, &j)
}

func (j *JobList) JsonMarshal() ([]byte, error) {
	return json.Marshal(j)
}

func (j *JobList) AddItemFromStr(objectStr string) error {
	object := &Job{}
	buf, err := strconv.Unquote(objectStr)
	err = object.JsonUnmarshal([]byte(buf))
	if err != nil {
		return err
	}
	j.Items = append(j.Items, *object)
	return nil
}

func (j *JobList) AppendItemsFromStr(objectStrs []string) error {
	for _, obj := range objectStrs {
		object := &Job{}
		err := object.JsonUnmarshal([]byte(obj))
		if err != nil {
			return err
		}
		j.Items = append(j.Items, *object)
	}
	return nil
}

func (j *JobList) GetItems() any {
	return j.Items
}

func (j *JobList) GetIApiObjectArr() (res []IApiObject) {
	for _, item := range j.Items {
		itemTemp := item
		res = append(res, &itemTemp)
	}
	return res
}

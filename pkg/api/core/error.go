package core

import "minik8s/pkg/api/types"

// ErrorApiObject ApiObject for error reporting
type ErrorApiObject struct {
	err error
	msg string
}

func (e *ErrorApiObject) GetError() error {
	return e.err
}

func (e *ErrorApiObject) SetError(err error) {
	e.err = err
}

func (e *ErrorApiObject) SetMsg(msg string) {
	e.msg = msg
}

func (e *ErrorApiObject) SetUID(uid types.UID) {
	panic("ErrorApiObject: this method should not be called!")
}

func (e *ErrorApiObject) GetUID() types.UID {
	panic("ErrorApiObject: this method should not be called!")
}

func (e *ErrorApiObject) JsonUnmarshal(data []byte) error {
	panic("ErrorApiObject: this method should not be called!")
}

func (e *ErrorApiObject) JsonMarshal() ([]byte, error) {
	panic("ErrorApiObject: this method should not be called!")
}

func (e *ErrorApiObject) JsonUnmarshalStatus(data []byte) error {
	panic("ErrorApiObject: this method should not be called!")
}

func (e *ErrorApiObject) JsonMarshalStatus() ([]byte, error) {
	panic("ErrorApiObject: this method should not be called!")
}

func (e *ErrorApiObject) SetStatus(s IApiObjectStatus) bool {
	panic("ErrorApiObject: this method should not be called!")
}

func (e *ErrorApiObject) GetStatus() IApiObjectStatus {
	panic("ErrorApiObject: this method should not be called!")
}

func (e *ErrorApiObject) GetResourceVersion() string {
	panic("ErrorApiObject: this method should not be called!")
}

func (e *ErrorApiObject) SetResourceVersion(version string) {
	panic("ErrorApiObject: this method should not be called!")
}

func (e *ErrorApiObject) CreateFromEtcdString(str string) error {
	panic("ErrorApiObject: this method should not be called!")
}

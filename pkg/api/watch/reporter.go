package watch

import (
	"log"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/types"
)

// Reporter hides the details of how an error is turned into a runtime.Object for
// reporting on a watch stream since this package may not import a higher level report.
type Reporter interface {
	// AsObject must convert err into a valid runtime.Object for the watch stream.
	AsObject(err error) core.IApiObject
}

type DefaultReporter struct {
}

func NewDefaultReporter() *DefaultReporter {
	return &DefaultReporter{}
}

func (d *DefaultReporter) AsObject(err error) core.IApiObject {
	log.Printf(err.Error())
	obj := core.CreateApiObject(types.ErrorObjectType)
	errObj := obj.(*core.ErrorApiObject)
	errObj.SetError(err)
	errObj.SetMsg(err.Error())
	return errObj
}

package watch

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/types"
	"minik8s/pkg/apiserver/etcd"
)

// Decoder allows StreamWatcher to watch any stream for which a Decoder can be written.
type Decoder interface {
	// Decode should return the type of event, the decoded object, or an error.
	// An error will cause StreamWatcher to call Close(). Decode should block until
	// it has data or an error occurs.
	Decode() (event *Event, err error)

	// Close should close the underlying io.Reader, signalling to the source of
	// the stream that it is no longer being watched. Close() must cause any
	// outstanding call to Decode() to return with an error of some sort.
	Close()
}

type EtcdEventDecoder struct {
	respBody   io.ReadCloser
	source     *bufio.Reader
	objectType types.ApiObjectType
}

func NewEtcdEventDecoder(body io.ReadCloser, ty types.ApiObjectType) *EtcdEventDecoder {
	d := &EtcdEventDecoder{
		respBody:   body,
		source:     bufio.NewReader(body),
		objectType: ty,
	}
	return d
}

// ConvertEvent convert event buf client received in watch first to etcd.Event, and then to
// watch.Event step by step
func (e *EtcdEventDecoder) convertEvent(buf []byte, ty types.ApiObjectType) (*Event, error) {

	log.Printf("[EtcdEventDecoder][ConvertEvent] buf: %v\n", string(buf))
	event := &etcd.Event{}
	err := json.Unmarshal(buf, event)
	if err != nil {
		log.Printf("[EtcdEventDecoder][ConvertEvent] Unmarshal APIServer event failed: %v\n", err)
		return nil, err
	}

	newEvent := &Event{}
	newEvent.Object = core.CreateApiObject(ty)

	switch event.Type {
	case etcd.EventTypePut:

		err = newEvent.Object.JsonUnmarshal(event.Kv.Value)
		if err != nil {
			log.Printf("[EtcdEventDecoder][ConvertEvent] Event JsonUnmarshal failed\n")
			return nil, err
		}

		if event.Kv.CreateRevision == event.Kv.ModRevision {
			newEvent.Type = Added
		} else {
			newEvent.Type = Modified
		}

		newEvent.Version = event.Kv.Version
		newEvent.CreateRevision = event.Kv.CreateRevision

	case etcd.EventTypeDelete:
		newEvent.Type = Deleted
	}

	newEvent.Key = string(event.Kv.Key)
	newEvent.ModRevision = event.Kv.ModRevision

	return newEvent, nil
}

func (e *EtcdEventDecoder) Decode() (event *Event, err error) {
	buf, err := e.source.ReadBytes(byte(SeparationChar))

	if err != nil {
		log.Printf("[EtcdEventDecoder] Watch %v Error: %v\n", e.objectType, err)
		return nil, err
	}

	buf[len(buf)-1] = SeparationChar

	return e.convertEvent(buf, e.objectType)

	//log.Print(event)
	//log.Print(event.Object)
}

func (e *EtcdEventDecoder) Close() {
	err := e.respBody.Close()
	if err != nil {
		log.Printf("[EtcdEventDecoder] %v respBody.Close Error: %v\n", e.objectType, err)
		return
	}
}

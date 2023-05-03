package client

import (
	"fmt"
	"log"
	"minik8s/pkg/api/core"
	"minik8s/pkg/client/listwatch"
	"os"
	"testing"
)

func GetPodJsonFilename() string {
	return "../../examples/pod/metrics-server.json"
}

var name string

func TestPostPod(t *testing.T) {
	filename := GetPodJsonFilename()
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("os.ReadFile(filename) failed: ", err)
		return
	}

	newObject := core.CreateApiObject(core.PodObjectType)
	err = newObject.JsonUnmarshal(jsonData)
	if err != nil {
		fmt.Println("newObject.JsonUnmarshal failed: ", err)
		return
	}
	rc, _ := NewRESTClient(core.PodObjectType)
	code, resp, err := rc.Post(newObject)
	if err != nil {
		return
	}
	fmt.Printf("code: %v, resp %v\n", code, resp)
	name = resp.UID
	fmt.Printf("name %v\n", name)
}

func TestPutPod(t *testing.T) {
	filename := GetPodJsonFilename()
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("os.ReadFile(filename) failed: ", err)
		return
	}

	newObject := core.CreateApiObject(core.PodObjectType)
	err = newObject.JsonUnmarshal(jsonData)
	if err != nil {
		fmt.Println("newObject.JsonUnmarshal failed: ", err)
		return
	}
	rc, _ := NewRESTClient(core.PodObjectType)
	fmt.Printf("name %v\n", name)
	code, resp, err := rc.Put(name, newObject)
	fmt.Printf("code: %v, resp %v\n", code, resp)
	if err != nil {
		return
	}
}

func TestGetPod(t *testing.T) {
	rc, _ := NewRESTClient(core.PodObjectType)
	res, err := rc.Get(name)
	if err != nil {
		return
	}
	fmt.Printf("resp %v\n", res)
}

func TestGetAllPod(t *testing.T) {
	rc, _ := NewRESTClient(core.PodObjectType)
	res, err := rc.GetAll()
	if err != nil {
		return
	}
	fmt.Printf("resp %v\n", res)
}

func TestDeletePod(t *testing.T) {
	rc, _ := NewRESTClient(core.PodObjectType)
	res, err := rc.Delete(name)
	if err != nil {
		return
	}
	fmt.Printf("resp %v\n", res)
}

func TestWatchAllPods(t *testing.T) {
	rc, err := NewRESTClient(core.PodObjectType)
	if err != nil {
		return
	}

	lw := listwatch.NewListWatchFromClient(rc)

	wi, err := lw.Watch()
	if err != nil {
		return
	}

	sum := 0
	for {
		if sum >= 3 {
			break
		} else {
			select {
			case e := <-wi.ResultChan():
				log.Printf("[RESTClient] event %v\n", e)
				log.Printf("[RESTClient] event object %v\n", e.Object)
				sum++
			}
		}
	}

	// stop watch
	wi.Stop()

}

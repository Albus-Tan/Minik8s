package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"log"
	"minik8s/pkg/api"
	"minik8s/pkg/api/core"
	"minik8s/pkg/apiserver/etcd"
	"net/http"
)

/*--------------------- Pod ---------------------*/
//	log.Printf(c.Request.URL.Path) // /api/pods/actual-name
//	log.Printf(c.FullPath())       // /api/pods/:name

func HandlePostPod(c *gin.Context) {

	// read request body
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	// parse request body from json to core.Pod type
	var newPod core.Pod
	err = json.Unmarshal(buf, &newPod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	// generate uuid for Pod
	uuidV4 := uuid.New()
	podUID := uuidV4.String()
	log.Printf("[apiserver] generate new Pod UID: %v", podUID)
	newPod.UID = podUID
	buf, err = json.Marshal(newPod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	// put Pod info into etcd
	err = etcd.Put(c.Request.URL.Path+podUID, string(buf))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "OK", "uid": podUID})
	}
}

func HandlePutPod(c *gin.Context) {
	// check if Pod exist
	has, err := etcd.Has(c.Request.URL.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{"status": "ERR", "error": "No such pod"})
		return
	}

	// read request body
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	// put/update Pod info into etcd
	err = etcd.Put(c.Request.URL.Path, string(buf))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	}

}

func HandleDeletePod(c *gin.Context) {
	// check if Pod exist
	has, err := etcd.Has(c.Request.URL.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{"status": "ERR", "error": "No such pod"})
		return
	}

	// delete Pod in etcd
	err = etcd.Delete(c.Request.URL.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	}
}

func HandleGetPod(c *gin.Context) {
	pod, err := etcd.Get(c.Request.URL.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else if pod == etcd.EmptyGetResult {
		c.JSON(http.StatusNotFound, gin.H{"status": "ERR", "error": "No such pod"})
	} else {
		c.JSON(http.StatusOK, pod)
	}
}

func HandleGetPods(c *gin.Context) {
	pods, err := etcd.GetAllWithPrefix(c.Request.URL.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, pods)
	}
}

func HandleWatchPod(c *gin.Context) {
	resourceURL := api.PodsURL + c.Param("name")

	// check if Pod exist
	has, err := etcd.Has(resourceURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{"status": "ERR", "error": "No such pod"})
		return
	}

	// register watch
	log.Printf("[apiserver][HandleWatchPod] Start watching resourceURL %v\n", resourceURL)
	cancel, ch := etcd.Watch(resourceURL)
	flusher, _ := c.Writer.(http.Flusher)
	for {
		select {
		case ev := <-ch:
			val, err := json.Marshal(string(ev.Kv.Value))
			if err != nil {
				log.Printf("[apiserver][HandleWatchPod] json parse error, cancel watch task\n")
				cancel()
				c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
				return
			}
			switch ev.Type {
			case etcd.EventTypeDelete:
				// cancel watch after delete
				log.Printf("[apiserver] Pod delete, cancel watch task\n")
				cancel()
				c.JSON(http.StatusOK, gin.H{"status": "OK"})
				return
			case etcd.EventTypePut:
				log.Printf("[apiserver] Pod put\n")
			default:
				// will not reach here
			}
			_, err = fmt.Fprintf(c.Writer, "%v\n", val)
			if err != nil {
				log.Printf("[apiserver][HandleWatchPod] fail to write to client, cancel watch task\n")
				cancel()
				c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
				return
			}
			flusher.Flush()
		case <-c.Request.Context().Done():
			log.Printf("[apiserver] Connection closed, cancel watch task\n")
			cancel()
			c.JSON(http.StatusOK, gin.H{"status": "OK"})
			return
		default:
			// when ch is blocked
		}
	}
}

func HandleWatchPods(c *gin.Context) {
	resourceURL := api.PodsURL

	// register watch
	log.Printf("[apiserver][HandleWatchPods] Start watching resourceURL %v\n", resourceURL)
	cancel, ch := etcd.WatchAllWithPrefix(resourceURL)
	flusher, _ := c.Writer.(http.Flusher)
	for {
		select {
		case ev := <-ch:
			val, err := json.Marshal(string(ev.Kv.Value))
			if err != nil {
				log.Printf("[apiserver][HandleWatchPods] json parse error, cancel watch task\n")
				cancel()
				c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
				return
			}
			switch ev.Type {
			case etcd.EventTypeDelete:
				// cancel watch after delete
				log.Printf("[apiserver] Pod delete\n")
			case etcd.EventTypePut:
				log.Printf("[apiserver] Pod put\n")
			default:
				// will not reach here
			}
			_, err = fmt.Fprintf(c.Writer, "%v\n", val)
			if err != nil {
				log.Printf("[apiserver][HandleWatchPods] fail to write to client, cancel watch task\n")
				cancel()
				c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
				return
			}
			flusher.Flush()
		case <-c.Request.Context().Done():
			log.Printf("[apiserver] Connection closed, cancel watch task\n")
			cancel()
			c.JSON(http.StatusOK, gin.H{"status": "OK"})
			return
		default:
			// when ch is blocked
		}
	}
}

func HandleGetPodStatus(c *gin.Context) {
	podJson, err := etcd.Get(api.PodsURL + c.Param("name"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else if podJson == etcd.EmptyGetResult {
		c.JSON(http.StatusNotFound, gin.H{"status": "ERR", "error": "No such podJson"})
	} else {
		pod := &core.Pod{}
		err = json.Unmarshal([]byte(podJson), &pod)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
			return
		}

		podStatus, err := json.Marshal(pod.Status)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, string(podStatus))
	}
}

func HandlePutPodStatus(c *gin.Context) {
	// check if Pod exist
	etcdURL := api.PodsURL + c.Param("name")
	has, err := etcd.Has(etcdURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{"status": "ERR", "error": "No such pod"})
		return
	}

	// read request body
	newStatus, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	// parse new status
	podStatus := &core.PodStatus{}
	err = json.Unmarshal(newStatus, &podStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	// read old pod
	podJson, err := etcd.Get(api.PodsURL + c.Param("name"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	// parse old pod
	pod := &core.Pod{}
	err = json.Unmarshal([]byte(podJson), &pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	// update status
	pod.Status = *podStatus

	// marshal new pod
	buf, err := json.Marshal(pod)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	// put/update Pod info into etcd
	err = etcd.Put(etcdURL, string(buf))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	}
}

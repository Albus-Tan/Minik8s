package handlers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"minik8s/pkg/api"
	"minik8s/pkg/api/core"
	"minik8s/pkg/apiserver/etcd"
	"net/http"
)

/*--------------------- Pod ---------------------*/
//	log.Printf(c.Request.URL.Path) // /api/pods/actual-name
//	log.Printf(c.FullPath())       // /api/pods/:name

func HandlePostPod(c *gin.Context) {
	handlePostObject(c, core.PodObjectType)
}

func HandlePutPod(c *gin.Context) {
	handlePutObject(c, core.PodObjectType)
}

func HandleDeletePod(c *gin.Context) {
	handleDeleteObject(c, core.PodObjectType)
}

func HandleGetPod(c *gin.Context) {
	handleGetObject(c, core.PodObjectType)
}

func HandleGetPods(c *gin.Context) {
	handleGetObjects(c, core.PodObjectType)
}

func HandleWatchPod(c *gin.Context) {
	resourceURL := api.PodsURL + c.Param("name")
	handleWatchObject(c, core.PodObjectType, resourceURL)
}

func HandleWatchPods(c *gin.Context) {
	resourceURL := api.PodsURL
	handleWatchObjects(c, core.PodObjectType, resourceURL)
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

package handlers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"log"
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
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(200, gin.H{"status": "OK", "uid": podUID})
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
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(200, gin.H{"status": "OK"})
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
		c.JSON(500, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(200, gin.H{"status": "OK"})
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

}

func HandleWatchPods(c *gin.Context) {

}

func HandleGetPodStatus(c *gin.Context) {

}

func HandlePutPodStatus(c *gin.Context) {

}

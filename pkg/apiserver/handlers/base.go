package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"log"
	"minik8s/pkg/api/core"
	"minik8s/pkg/apiserver/etcd"
	"net/http"
)

func handlePostObject(c *gin.Context, ty core.ApiObjectType) {

	// read request body
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	// parse request body from json to core.{ApiObject} type
	newObject := core.CreateApiObject(ty)
	err = newObject.JsonUnmarshal(buf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	//log.Printf("[apiserver] JsonUnmarshal buf %v", string(buf))

	// generate uuid for {ApiObject}
	uuidV4 := uuid.New()
	objectUID := uuidV4.String()
	log.Printf("[apiserver] generate new %v UID: %v", ty, objectUID)
	newObject.SetUID(objectUID)
	buf, err = newObject.JsonMarshal()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	//log.Printf("[apiserver] JsonMarshal buf %v", string(buf))

	// put {ApiObject} info into etcd
	err = etcd.Put(c.Request.URL.Path+objectUID, string(buf))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "OK", "uid": objectUID})
	}
}

func handlePutObject(c *gin.Context, ty core.ApiObjectType) {
	// check if {ApiObject} exist
	has, err := etcd.Has(c.Request.URL.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{"status": "ERR", "error": fmt.Sprintf("No such %v", ty)})
		return
	}

	// read request body
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	// put/update {ApiObject} info into etcd
	err = etcd.Put(c.Request.URL.Path, string(buf))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	}
}

func handleDeleteObject(c *gin.Context, ty core.ApiObjectType) {
	// check if {ApiObject} exist
	has, err := etcd.Has(c.Request.URL.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{"status": "ERR", "error": fmt.Sprintf("No such %v", ty)})
		return
	}

	// delete {ApiObject} in etcd
	err = etcd.Delete(c.Request.URL.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	}
}

func handleGetObject(c *gin.Context, ty core.ApiObjectType) {
	object, err := etcd.Get(c.Request.URL.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else if object == etcd.EmptyGetResult {
		c.JSON(http.StatusNotFound, gin.H{"status": "ERR", "error": fmt.Sprintf("No such %v", ty)})
	} else {
		c.JSON(http.StatusOK, object)
	}
}

func handleGetObjects(c *gin.Context, ty core.ApiObjectType) {
	objects, err := etcd.GetAllWithPrefix(c.Request.URL.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, objects)
	}
}

func handleWatchObject(c *gin.Context, ty core.ApiObjectType, resourceURL string) {
	// check if {ApiObject} exist
	has, err := etcd.Has(resourceURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{"status": "ERR", "error": fmt.Sprintf("No such %v", ty)})
		return
	}

	// register watch
	log.Printf("[apiserver][HandleWatch%v] Start watching resourceURL %v\n", ty, resourceURL)
	cancel, ch := etcd.Watch(resourceURL)
	flusher, _ := c.Writer.(http.Flusher)
	for {
		select {
		case ev := <-ch:
			data, err := json.Marshal(string(ev.Kv.Value))
			val := string(data)
			if err != nil {
				log.Printf("[apiserver][HandleWatch%v] json parse error, cancel watch task\n", ty)
				cancel()
				c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
				return
			}
			switch ev.Type {
			case etcd.EventTypeDelete:
				// cancel watch after delete
				log.Printf("[apiserver] %v delete, cancel watch task\n", ty)
				cancel()
				c.JSON(http.StatusOK, gin.H{"status": "OK"})
				return
			case etcd.EventTypePut:
				log.Printf("[apiserver] %v put\n", ty)
			default:
				// will not reach here
			}
			_, err = fmt.Fprintf(c.Writer, "%v\n", val)
			if err != nil {
				log.Printf("[apiserver][HandleWatch%v] fail to write to client, cancel watch task\n", ty)
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

func handleWatchObjects(c *gin.Context, ty core.ApiObjectType, resourceURL string) {
	// register watch
	log.Printf("[apiserver][HandleWatch%vs] Start watching resourceURL %v\n", ty, resourceURL)
	cancel, ch := etcd.WatchAllWithPrefix(resourceURL)
	flusher, _ := c.Writer.(http.Flusher)
	for {
		select {
		case ev := <-ch:
			data, err := json.Marshal(string(ev.Kv.Value))
			val := string(data)
			if err != nil {
				log.Printf("[apiserver][HandleWatch%vs] json parse error, cancel watch task\n", ty)
				cancel()
				c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
				return
			}
			switch ev.Type {
			case etcd.EventTypeDelete:
				// cancel watch after delete
				log.Printf("[apiserver] %v delete\n", ty)
			case etcd.EventTypePut:
				log.Printf("[apiserver] %v put\n", ty)
			default:
				// will not reach here
			}
			_, err = fmt.Fprintf(c.Writer, "%v\n", val)
			if err != nil {
				log.Printf("[apiserver][HandleWatch%vs] fail to write to client, cancel watch task\n", ty)
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

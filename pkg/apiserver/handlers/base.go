package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/types"
	"minik8s/pkg/apiserver/etcd"
	"minik8s/pkg/logger"
	"minik8s/utils"
	"net/http"
)

func HandleClearAll(c *gin.Context) {
	err := etcd.Clear()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "OK"})
}

func handlePostObject(c *gin.Context, ty types.ApiObjectType) {

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
	//logger.ApiServerLogger.Printf("[apiserver] JsonUnmarshal buf %v", string(buf))

	// generate uuid for {ApiObject}
	objectUID := utils.GenerateUID()
	logger.ApiServerLogger.Printf("[apiserver] generate new %v UID: %v", ty, objectUID)
	newObject.SetUID(objectUID)

	// lock for version get, set and store
	etcd.VLock.Lock()
	defer etcd.VLock.Unlock()

	// set object ResourceVersion
	createVersion := etcd.Rvm.GetNextResourceVersion()
	newObject.SetResourceVersion(createVersion)

	buf, err = newObject.JsonMarshal()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	//logger.ApiServerLogger.Printf("[apiserver] JsonMarshal buf %v", string(buf))

	etcdPath := c.Request.URL.Path
	if ty == types.FuncTemplateObjectType {
		f := newObject.(*core.Func)
		etcdPath += f.Spec.Name
	} else {
		etcdPath += objectUID
	}

	// process dns config
	if ty == types.DnsObjectType {
		dns := newObject.(*core.DNS)
		handleAddCoreDnsConfig(dns)
	}

	// put/update {ApiObject} info into etcd
	err, newVersion := etcd.Put(etcdPath, string(buf))
	logger.ApiServerLogger.Printf("[apiserver] generate new %v: json ResourceVersion %v, current ResourceVersion %v", ty, createVersion, newVersion)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "OK", "uid": objectUID, "resourceVersion": createVersion})
	}
}

func handlePutObject(c *gin.Context, ty types.ApiObjectType) {
	// check if {ApiObject} exist
	has, versionHas, err := etcd.HasWithVersion(c.Request.URL.Path)
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

	// parse request body from json to core.{ApiObject} type
	newObject := core.CreateApiObject(ty)
	err = newObject.JsonUnmarshal(buf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	// get object old version
	oldVersion := newObject.GetResourceVersion()
	if versionHas != oldVersion {
		c.JSON(http.StatusConflict, gin.H{"status": "FAILED", "error": fmt.Sprintf("Old version %v unmatch current version %v, %v has been modified by others, please GET for the new version and retry PUT operation", oldVersion, versionHas, ty)})
		return
	}

	// lock for version get, set and store
	etcd.VLock.Lock()
	defer etcd.VLock.Unlock()

	// update object new version
	newObject.SetResourceVersion(etcd.Rvm.GetNextResourceVersion())

	buf, err = newObject.JsonMarshal()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	// put/update {ApiObject} info into etcd
	err, newVersion, success := etcd.CheckVersionPut(c.Request.URL.Path, string(buf), oldVersion)
	if !success {
		c.JSON(http.StatusConflict, gin.H{"status": "FAILED", "error": fmt.Sprintf("Old version unmatch current version, %v has been modified by others, please GET for the new version and retry PUT operation", ty)})
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "OK", "resourceVersion": newVersion})
	}
}

func handleDeleteObject(c *gin.Context, ty types.ApiObjectType) {
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

	// process dns config
	if ty == types.DnsObjectType {
		objectStr, _ := etcd.Get(c.Request.URL.Path)
		newObject := core.CreateApiObject(ty)
		err = newObject.CreateFromEtcdString(objectStr)
		dns := newObject.(*core.DNS)
		handleDeleteCoreDnsConfig(dns)
	}

	// delete {ApiObject} in etcd
	err = etcd.Delete(c.Request.URL.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "OK"})
	}
}

func handleGetObject(c *gin.Context, ty types.ApiObjectType) {
	objectStr, err := etcd.Get(c.Request.URL.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else if objectStr == etcd.EmptyGetResult {
		c.JSON(http.StatusNotFound, gin.H{"status": "ERR", "error": fmt.Sprintf("No such %v", ty)})
	} else {
		object := core.CreateApiObject(ty)
		err = object.CreateFromEtcdString(objectStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			c.JSON(http.StatusOK, object)
		}
	}
}

func handleGetObjects(c *gin.Context, ty types.ApiObjectType) {
	objects, err := etcd.GetAllWithPrefix(c.Request.URL.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		objectList := core.CreateApiObjectList(ty)
		err := objectList.AppendItemsFromStr(objects)
		if err != nil {
			logger.ApiServerLogger.Println("[apiserver] handleGetObjects objectList.AppendItemsFromStr failed", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, objectList)
	}
}

func handleWatchObjectAndStatus(c *gin.Context, ty types.ApiObjectType, resourceURL string) {
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
	logger.ApiServerLogger.Printf("[apiserver][HandleWatch%v] Start watching resourceURL %v\n", ty, resourceURL)
	cancel, ch := etcd.WatchAllWithPrefix(resourceURL)
	flusher, _ := c.Writer.(http.Flusher)
	for {
		select {
		case ev := <-ch:
			switch ev.Type {
			case etcd.EventTypeDelete:
				event, err := json.Marshal(ev)
				if err != nil {
					logger.ApiServerLogger.Printf("[apiserver][HandleWatch%vs] json.Marshal event failed, cancel watch task\n", ty)
					cancel()
					c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
					return
				}
				_, err = fmt.Fprintf(c.Writer, "%v\n", string(event))
				if err != nil {
					logger.ApiServerLogger.Printf("[apiserver][HandleWatch%v] fail to write to client, cancel watch task\n", ty)
					cancel()
					c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
					return
				}
				flusher.Flush()

				// cancel watch after delete
				logger.ApiServerLogger.Printf("[apiserver] %v delete, cancel watch task\n", ty)
				cancel()

				c.JSON(http.StatusOK, gin.H{"status": "OK"})
				return
			case etcd.EventTypePut:
				logger.ApiServerLogger.Printf("[apiserver] %v put\n", ty)
			default:
				// will not reach here
			}

			event, err := json.Marshal(ev)
			if err != nil {
				logger.ApiServerLogger.Printf("[apiserver][HandleWatch%vs] json.Marshal event failed, cancel watch task\n", ty)
				cancel()
				c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
				return
			}
			_, err = fmt.Fprintf(c.Writer, "%v\n", string(event))

			if err != nil {
				logger.ApiServerLogger.Printf("[apiserver][HandleWatch%v] fail to write to client, cancel watch task\n", ty)
				cancel()
				c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
				return
			}
			flusher.Flush()
		case <-c.Request.Context().Done():
			logger.ApiServerLogger.Printf("[apiserver] Connection closed, cancel watch task\n")
			cancel()
			c.JSON(http.StatusOK, gin.H{"status": "OK"})
			return
		default:
			// when ch is blocked
		}
	}
}

func handleWatchObjectsAndStatus(c *gin.Context, ty types.ApiObjectType, resourceURL string) {
	// register watch
	logger.ApiServerLogger.Printf("[apiserver][HandleWatch%vs] Start watching resourceURL %v\n", ty, resourceURL)
	cancel, ch := etcd.WatchAllWithPrefix(resourceURL)
	flusher, _ := c.Writer.(http.Flusher)
	for {
		select {
		case ev := <-ch:
			switch ev.Type {
			case etcd.EventTypeDelete:
				logger.ApiServerLogger.Printf("[apiserver] %v delete\n", ty)
			case etcd.EventTypePut:
				logger.ApiServerLogger.Printf("[apiserver] %v put\n", ty)
			default:
				// will not reach here
			}
			event, err := json.Marshal(ev)
			if err != nil {
				logger.ApiServerLogger.Printf("[apiserver][HandleWatch%vs] json.Marshal event failed, cancel watch task\n", ty)
				cancel()
				c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
				return
			}
			_, err = fmt.Fprintf(c.Writer, "%v\n", string(event))

			if err != nil {
				logger.ApiServerLogger.Printf("[apiserver][HandleWatch%vs] fail to write to client, cancel watch task\n", ty)
				cancel()
				c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
				return
			}
			flusher.Flush()
		case <-c.Request.Context().Done():
			logger.ApiServerLogger.Printf("[apiserver][HandleWatch%vs] Connection closed, cancel watch task\n", ty)
			cancel()
			c.JSON(http.StatusOK, gin.H{"status": "OK"})
			return
		default:
			// when ch is blocked
		}
	}
}

func handleGetObjectStatus(c *gin.Context, ty types.ApiObjectType, resourceURL string) {
	objectJson, err := etcd.Get(resourceURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else if objectJson == etcd.EmptyGetResult {
		c.JSON(http.StatusNotFound, gin.H{"status": "ERR", "error": fmt.Sprintf("No such %v", ty)})
	} else {
		object := core.CreateApiObject(ty)
		err = object.JsonUnmarshal([]byte(objectJson))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, object.GetStatus())
	}
}

func handlePutObjectStatus(c *gin.Context, ty types.ApiObjectType, etcdURL string) {

	has, versionHas, err := etcd.HasWithVersion(etcdURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{"status": "ERR", "error": fmt.Sprintf("No such %v", ty)})
		return
	}

	// read request body
	newStatus, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	// parse new status
	objectStatus := core.CreateApiObjectStatus(ty)
	err = objectStatus.JsonUnmarshal(newStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	// read old {ApiObject}
	objectJson, err := etcd.Get(etcdURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	// parse old {ApiObject}
	object := core.CreateApiObject(ty)
	err = object.JsonUnmarshal([]byte(objectJson))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	// update status
	if success := object.SetStatus(objectStatus); !success {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": fmt.Sprintf("%vStatus update error, type unmatch", ty)})
		return
	}

	// get object old version
	oldVersion := object.GetResourceVersion()
	if versionHas != oldVersion {
		c.JSON(http.StatusConflict, gin.H{"status": "FAILED", "error": fmt.Sprintf("Old version unmatch current version, %v has been modified by others, please GET for the new version and retry PUT operation", ty)})
		return
	}

	// lock for version get, set and store
	etcd.VLock.Lock()
	defer etcd.VLock.Unlock()

	// update object new version
	object.SetResourceVersion(etcd.Rvm.GetNextResourceVersion())

	// marshal new object
	buf, err := object.JsonMarshal()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	// put/update {ApiObject} info into etcd
	err, newVersion, success := etcd.CheckVersionPut(etcdURL, string(buf), oldVersion)
	if !success {
		c.JSON(http.StatusConflict, gin.H{"status": "FAILED", "error": fmt.Sprintf("Old version unmatch current version, %v has been modified by others, please GET for the new version and retry PUT operation", ty)})
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{"status": "OK", "resourceVersion": newVersion})
	}
}

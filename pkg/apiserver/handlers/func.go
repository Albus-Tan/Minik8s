package handlers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"minik8s/config"
	"minik8s/pkg/api"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/generate"
	"minik8s/pkg/api/types"
	"minik8s/pkg/apiserver/etcd"
	"minik8s/pkg/logger"
	"minik8s/utils"
	"net/http"
	"time"
)

/*--------------------- FuncTemplate ---------------------*/

func HandlePostFuncTemplate(c *gin.Context) {
	handlePostObject(c, types.FuncTemplateObjectType)
}

func HandlePutFuncTemplate(c *gin.Context) {
	handlePutObject(c, types.FuncTemplateObjectType)
}

func HandleDeleteFuncTemplate(c *gin.Context) {
	handleDeleteObject(c, types.FuncTemplateObjectType)
}

func HandleGetFuncTemplate(c *gin.Context) {
	handleGetObject(c, types.FuncTemplateObjectType)
}

func HandleGetFuncTemplates(c *gin.Context) {
	handleGetObjects(c, types.FuncTemplateObjectType)
}

func HandleWatchFuncTemplate(c *gin.Context) {
	resourceURL := api.FuncTemplatesURL + c.Param("name")
	handleWatchObjectAndStatus(c, types.FuncTemplateObjectType, resourceURL)
}

func HandleWatchFuncTemplates(c *gin.Context) {
	resourceURL := api.FuncTemplatesURL
	handleWatchObjectsAndStatus(c, types.FuncTemplateObjectType, resourceURL)
}

func HandleGetFuncTemplateStatus(c *gin.Context) {
	resourceURL := api.FuncTemplatesURL + c.Param("name")
	handleGetObjectStatus(c, types.FuncTemplateObjectType, resourceURL)
}

func HandlePutFuncTemplateStatus(c *gin.Context) {
	etcdURL := api.FuncTemplatesURL + c.Param("name")
	handlePutObjectStatus(c, types.FuncTemplateObjectType, etcdURL)
}

/*--------------------- Function Instance (For User) ---------------------*/

// HandleFuncCall Create a Function Instance (run function)
// return id (instance func id)
// POST /api/funcs/{name}
func HandleFuncCall(c *gin.Context) {

	// Read request body
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	logger.ApiServerLogger.Printf("HandleFuncCall body: %v\n", string(buf))

	// Get func template of name
	funcTemplate, err := getFuncTemplate(c)
	if err != nil {
		logger.ApiServerLogger.Printf("HandleFuncCall err: %v\n", err)
		return
	}

	// Generate instanceId for current call (unique among all calls)
	instanceId := utils.GenerateInstanceID()

	// call function
	// request body used as function args
	doInsideFuncCall(instanceId, funcTemplate, string(buf), c)

	// return instanceId
	c.JSON(http.StatusOK, gin.H{"status": "OK", "id": instanceId})
}

// HandleGetResult Get the function result
// GET /api/funcs/{id}
func HandleGetResult(c *gin.Context) {
	instanceId := c.Param("id")
	// Get result in etcd (key: FuncResultURL)
	value, err := etcd.Get(c.Request.URL.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else if value == etcd.EmptyGetResult {
		c.JSON(http.StatusNotFound, gin.H{"status": "ERR", "error": fmt.Sprintf("No result for instanceId %v now", instanceId)})
	} else {
		c.JSON(http.StatusOK, value)
	}
}

/*--------------------- Function Instance (For Impl) ---------------------*/

// HandleInsideFuncCall Create a Function Instance (run function), id is the instance func id called by user
// PUT /api/funcs/{name}/{id}
func HandleInsideFuncCall(c *gin.Context) {

	// Get instance id
	instanceId := c.Param("id")

	// Read request body
	buf, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return
	}

	logger.ApiServerLogger.Printf("HandleInsideFuncCall body: %v\n", string(buf))

	// Get functionName
	functionName := c.Param("name")
	if functionName == api.ReturnPreservedName {

		// Put function result into etcd

		// store result in etcd (key: FuncResultURL)
		etcdPath := api.FuncURL + instanceId
		// put/update result info into etcd
		// request body used as function result
		err, _ = etcd.Put(etcdPath, string(buf))

		logger.ApiServerLogger.Printf("HandleInsideFuncCall result for instanceId %v: %v", instanceId, string(buf))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
			return
		}
		// return instanceId
		c.JSON(http.StatusOK, gin.H{"status": "RET", "id": instanceId})
		return
	}

	// Get func template of name
	funcTemplate, err := getFuncTemplate(c)
	if err != nil {
		logger.ApiServerLogger.Printf("HandleInsideFuncCall err: %v\n", err)
		return
	}

	// call function
	// request body used as function args
	doInsideFuncCall(instanceId, funcTemplate, string(buf), c)

	// return instanceId
	c.JSON(http.StatusOK, gin.H{"status": "OK", "id": instanceId})

}

func updateFuncTemplate(c *gin.Context, funcTemplate *core.Func) error {

	resourceURL := api.FuncTemplatesURL + c.Param("name")

	// check if FuncTemplate exist
	has, versionHas, err := etcd.HasWithVersion(resourceURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return err
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{"status": "ERR", "error": fmt.Sprintf("No such func template")})
		return err
	}

	// get object old version
	oldVersion := funcTemplate.GetResourceVersion()
	if versionHas != oldVersion {
		c.JSON(http.StatusConflict, gin.H{"status": "FAILED", "error": fmt.Sprintf("Old version %v unmatch current version %v, func template has been modified by others, please GET for the new version and retry PUT operation", oldVersion, versionHas)})
		return err
	}

	// lock for version get, set and store
	etcd.VLock.Lock()
	defer etcd.VLock.Unlock()

	// update object new version
	funcTemplate.SetResourceVersion(etcd.Rvm.GetNextResourceVersion())

	buf, err := funcTemplate.JsonMarshal()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		return err
	}

	// put/update {ApiObject} info into etcd
	err, _, success := etcd.CheckVersionPut(resourceURL, string(buf), oldVersion)
	if !success {
		c.JSON(http.StatusConflict, gin.H{"status": "FAILED", "error": fmt.Sprintf("Old version unmatch current version, func template has been modified by others, please GET for the new version and retry PUT operation")})
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	}

	return nil
}

func getFuncTemplate(c *gin.Context) (funcTemplate *core.Func, err error) {
	resourceURL := api.FuncTemplatesURL + c.Param("name")
	objectStr, err := etcd.Get(resourceURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
	} else if objectStr == etcd.EmptyGetResult {
		c.JSON(http.StatusNotFound, gin.H{"status": "ERR", "error": fmt.Sprintf("No such %v template", types.FuncTemplateObjectType)})
	} else {
		object := core.CreateApiObject(types.FuncTemplateObjectType)
		err = object.CreateFromEtcdString(objectStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"status": "ERR", "error": err.Error()})
		} else {
			funcTemplate = object.(*core.Func)
			return funcTemplate, nil
		}
	}
	return nil, errors.New("getFuncTemplate failed\n")
}

// createPod directly create new pod, and add it to etcd
// The pod is not guaranteed running after the function return,
// it will be scheduled and then run by kubelet.
// Used for serverless v1
func createPod(newPod *core.Pod, objectUID types.UID) (resourceVersion string, err error) {
	newPod.SetUID(objectUID)
	etcd.VLock.Lock()
	defer etcd.VLock.Unlock()
	// set object ResourceVersion
	createVersion := etcd.Rvm.GetNextResourceVersion()
	newPod.SetResourceVersion(createVersion)

	buf, err := newPod.JsonMarshal()
	if err != nil {
		return "", err
	}

	etcdPath := api.PodsURL + objectUID

	// put/update Pod info into etcd
	err, newVersion := etcd.Put(etcdPath, string(buf))
	logger.ApiServerLogger.Printf("[apiserver] generate new Func Pod: json ResourceVersion %v, current ResourceVersion %v", createVersion, newVersion)
	if err != nil {
		return "", err
	} else {
		return createVersion, err
	}
}

// Used for serverless v2
func doInsideFuncCall(instanceId string, funcTemplate *core.Func, args string, c *gin.Context) {
	switch funcTemplate.TypeMeta.APIVersion {
	case "v1":
		doInsideFuncCallV1(instanceId, funcTemplate, args)
	case "v2":
		doInsideFuncCallV2(instanceId, funcTemplate, args, c)
	}
}

// Used for serverless v2
func doInsideFuncCallV2(instanceId string, funcTemplate *core.Func, args string, c *gin.Context) {

	// modify timestamp and counter in status
	originNum := funcTemplate.Status.Counter
	instanceNum := funcTemplate.Status.Counter
	maxInstanceNum := config.FuncDefaultMaxInstanceNum
	if funcTemplate.Spec.MaxInstanceNum != nil {
		maxInstanceNum = *funcTemplate.Spec.MaxInstanceNum
	}
	if instanceNum < maxInstanceNum {
		instanceNum += 1
	}
	funcTemplate.Status.Counter = instanceNum
	funcTemplate.Status.TimeStamp = time.Now()

	err := updateFuncTemplate(c, funcTemplate)
	if err != nil {
		logger.ApiServerLogger.Printf(
			"[apiserver] updateFuncTemplate failed: for instanceId %v of func %v, err: %v",
			instanceId, funcTemplate.Spec.Name, err)
		return
	}

	if originNum == 0 {
		// if there were no instance before, wait some time for instance cold boot
		time.Sleep(config.FuncCallColdBootWait)
	}

	// TODO @wjr for serverless v2, redirect http request to service,
	// 	use loop to wait for pod running

	logger.ApiServerLogger.Printf(
		"[apiserver] doInsideFuncCall v2 success: for instanceId %v, redirect to service UID %v of func %v",
		instanceId, funcTemplate.Status.ServiceUID, funcTemplate.Spec.Name)

}

// Used for serverless v1
func doInsideFuncCallV1(instanceId string, funcTemplate *core.Func, args string) {

	newPod := generate.EmptyPod()

	// generate uuid for Pod
	objectUID := utils.GenerateUID()
	logger.ApiServerLogger.Printf("[apiserver] generate new Func Pod UID: %v", objectUID)

	// TODO: @wjr fill in pod field by funcTemplate
	// 	such as Containers, Object Meta Name etc.
	newPod.Name = objectUID + "-" + funcTemplate.Name
	newPod.Spec = core.PodSpec{
		Containers: []core.Container{
			{
				Name:  "instance",
				Image: "lwsg/func-runner:0.10",
				Env: []core.EnvVar{
					{
						Name:  "_API_SERVER",
						Value: "http://10.180.253.214:8080",
					},
					{
						Name:  "_PRE_RUN",
						Value: funcTemplate.Spec.PreRun,
					},
					{
						Name:  "_UID",
						Value: instanceId,
					},
					{
						Name:  "_ARG",
						Value: args,
					},
					{
						Name:  "_FUNC",
						Value: funcTemplate.Spec.Function,
					},
					{
						Name:  "_LEFT",
						Value: funcTemplate.Spec.Left,
					},
					{
						Name:  "_RIGHT",
						Value: funcTemplate.Spec.Right,
					},
					{
						Name:  "_SELF",
						Value: objectUID,
					},
				},
			},
		},
		RestartPolicy: core.RestartPolicyNever,
	}

	// Used for serverless v1
	// create pod
	_, err := createPod(newPod, objectUID)
	if err != nil {
		return
	}

	logger.ApiServerLogger.Printf(
		"[apiserver] doInsideFuncCall v1: for instanceId %v, create pod UID %v of func %v",
		instanceId, objectUID, funcTemplate.Spec.Name)
}

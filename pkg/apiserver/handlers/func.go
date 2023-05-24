package handlers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"minik8s/pkg/api"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/generate"
	"minik8s/pkg/api/types"
	"minik8s/pkg/apiserver/etcd"
	"minik8s/pkg/logger"
	"minik8s/utils"
	"net/http"
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
	doInsideFuncCall(instanceId, funcTemplate, string(buf))

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
	doInsideFuncCall(instanceId, funcTemplate, string(buf))

	// return instanceId
	c.JSON(http.StatusOK, gin.H{"status": "OK", "id": instanceId})

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

func doInsideFuncCall(instanceId string, funcTemplate *core.Func, args string) {

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
				Image: "lwsg/func-runner:0.7",
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

	// create pod
	_, err := createPod(newPod, objectUID)
	if err != nil {
		return
	}

	logger.ApiServerLogger.Printf(
		"[apiserver] doInsideFuncCall: for instanceId %v, create pod UID %v of func %v",
		instanceId, objectUID, funcTemplate.Spec.Name)
}

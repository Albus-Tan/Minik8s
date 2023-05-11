package apiclient

import (
	"errors"
	"io"
	"minik8s/config"
	"minik8s/pkg/api"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/types"
	"minik8s/pkg/api/watch"
	httpclient "minik8s/pkg/apiclient/http"
	"minik8s/pkg/logger"
	"net/http"
	"time"
)

const HttpStatusNotSend = 0
const ReconnectInterval = 5 // watch reconnect seconds

// RESTClient This client performs generic REST functions such as Get, Put, Post, and Delete on specified paths.
type RESTClient struct {
	apiServerURL string // url of apiServer
	resourceURL  string // url of resource in pkg api
	resourceType types.ApiObjectType
}

// NewRESTClient creates a new RESTClient. This client performs generic REST functions
// such as Get, Put, Post, and Delete on specified paths.
// Different type of resource need to use different RESTClient
func NewRESTClient(ty types.ApiObjectType) (*RESTClient, error) {
	return &RESTClient{
		resourceType: ty,
		resourceURL:  core.GetApiObjectsURL(ty),
		apiServerURL: config.ApiServerUrl(),
	}, nil
}

func (c *RESTClient) URL() string {
	return c.apiServerURL + c.resourceURL
}

func (c *RESTClient) WatchURL() string {
	return c.apiServerURL + core.GetWatchApiObjectsURL(c.resourceType)
}

func (c *RESTClient) createApiObject() core.IApiObject {
	return core.CreateApiObject(c.resourceType)
}

func (c *RESTClient) createApiObjectStatus() core.IApiObjectStatus {
	return core.CreateApiObjectStatus(c.resourceType)
}

// Post begins a POST request.
func (c *RESTClient) Post(object core.IApiObject) (int, *api.PostResponse, error) {
	resourceURL := c.URL()
	content, err := object.JsonMarshal()
	if err != nil {
		logger.ApiClientLogger.Println("[RESTClient] http.Post JsonMarshal failed", err)
		return HttpStatusNotSend, nil, err
	}

	resp, err := httpclient.PostBytes(resourceURL, content)
	defer resp.Body.Close()
	if err != nil {
		logger.ApiClientLogger.Println("[RESTClient] http.Post failed", err)
		return HttpStatusNotSend, nil, err
	}

	postResp := &api.PostResponse{}
	err = postResp.FillResponse(resp)
	if err != nil {
		return resp.StatusCode, nil, err
	}

	if resp.StatusCode == http.StatusOK {
		return resp.StatusCode, postResp, nil
	} else {
		logger.ApiClientLogger.Println("[RESTClient] http.Post StatusCode not http.StatusOK", err)
		return resp.StatusCode, postResp, errors.New("StatusCode not 200")
	}
}

// Put begins a PUT request.
func (c *RESTClient) Put(name string, object core.IApiObject) (int, *api.PutResponse, error) {
	resourceURL := c.URL() + name
	content, err := object.JsonMarshal()
	if err != nil {
		logger.ApiClientLogger.Println("[RESTClient] http.Put JsonMarshal failed", err)
		return HttpStatusNotSend, nil, err
	}

	resp, err := httpclient.PutBytes(resourceURL, content)
	if err != nil {
		logger.ApiClientLogger.Println("[RESTClient] http.Put failed", err)
		return HttpStatusNotSend, nil, err
	}

	putResp := &api.PutResponse{}
	err = putResp.FillResponse(resp)
	if err != nil {
		return resp.StatusCode, nil, err
	}

	if resp.StatusCode == http.StatusOK {
		return resp.StatusCode, putResp, nil
	} else {
		logger.ApiClientLogger.Println("[RESTClient] http.Put StatusCode not http.StatusOK", err)
		return resp.StatusCode, putResp, errors.New("StatusCode not 200")
	}
}

// Get begins a GET request.
func (c *RESTClient) Get(name string) (core.IApiObject, error) {
	resourceURL := c.URL() + name
	object := c.createApiObject()

	resp, err := http.Get(resourceURL)
	if err != nil {
		logger.ApiClientLogger.Println("[RESTClient] http.Get failed", err)
		return nil, err
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ApiClientLogger.Println("[RESTClient] http.Get response io.ReadAll(resp.Body) failed", err)
		return nil, err
	}

	err = object.JsonUnmarshal(content)
	if err != nil {
		logger.ApiClientLogger.Printf("[RESTClient] http.Get response json.Unmarshal failed, err %v\n", err)
		return nil, err
	}

	return object, nil
}

func (c *RESTClient) GetStatus(name string) (core.IApiObjectStatus, error) {
	resourceURL := c.URL() + name + api.StatusSuffix
	objectStatus := c.createApiObjectStatus()

	resp, err := http.Get(resourceURL)
	if err != nil {
		logger.ApiClientLogger.Println("[RESTClient] http.GetStatus failed", err)
		return nil, err
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ApiClientLogger.Println("[RESTClient] http.GetStatus response io.ReadAll(resp.Body) failed", err)
		return nil, err
	}

	err = objectStatus.JsonUnmarshal(content)
	if err != nil {
		logger.ApiClientLogger.Printf("[RESTClient] http.GetStatus response json.Unmarshal failed, err %v\n", err)
		return nil, err
	}

	return objectStatus, nil
}

func (c *RESTClient) PutStatus(name string, object core.IApiObjectStatus) (int, *api.PutResponse, error) {
	resourceURL := c.URL() + name + api.StatusSuffix
	content, err := object.JsonMarshal()
	if err != nil {
		logger.ApiClientLogger.Println("[RESTClient] http.PutStatus JsonMarshal failed", err)
		return HttpStatusNotSend, nil, err
	}

	resp, err := httpclient.PutBytes(resourceURL, content)
	if err != nil {
		logger.ApiClientLogger.Println("[RESTClient] http.PutStatus failed", err)
		return HttpStatusNotSend, nil, err
	}

	putResp := &api.PutResponse{}
	err = putResp.FillResponse(resp)
	if err != nil {
		return resp.StatusCode, nil, err
	}

	if resp.StatusCode == http.StatusOK {
		return resp.StatusCode, putResp, nil
	} else {
		logger.ApiClientLogger.Println("[RESTClient] http.PutStatus StatusCode not http.StatusOK", err)
		return resp.StatusCode, putResp, errors.New("StatusCode not 200")
	}
}

func (c *RESTClient) GetAll() (objectList core.IApiObjectList, err error) {
	resourceURL := c.URL()

	resp, err := http.Get(resourceURL)
	if err != nil {
		logger.ApiClientLogger.Println("[RESTClient] http.GetAll failed", err)
		return nil, err
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.ApiClientLogger.Println("[RESTClient] http.GetAll response io.ReadAll(resp.Body) failed", err)
		return nil, err
	}

	objectList = core.CreateApiObjectList(c.resourceType)
	if len(content) == 0 {
		return objectList, nil
	}

	err = objectList.JsonUnmarshal(content)
	if err != nil {
		logger.ApiClientLogger.Printf("[RESTClient] http.Get response json.Unmarshal failed, err %v\n", err)
		return nil, err
	}

	if err != nil {
		logger.ApiClientLogger.Println("[RESTClient] http.GetAll response json.Unmarshal objects list failed", err)
		return nil, err
	}

	return objectList, nil
}

// Delete begins a DELETE request.
func (c *RESTClient) Delete(name string) (int, *api.DeleteResponse, error) {
	resourceURL := c.URL() + name

	cli := &http.Client{}
	req, err := http.NewRequest(http.MethodDelete, resourceURL, nil)
	if err != nil {
		logger.ApiClientLogger.Println("[RESTClient] http.Delete NewRequest create failed", err)
		return HttpStatusNotSend, nil, err
	}

	resp, err := cli.Do(req)
	if err != nil {
		logger.ApiClientLogger.Println("[RESTClient] http.Delete request send failed", err)
		return HttpStatusNotSend, nil, err
	}
	defer resp.Body.Close()

	delResp := &api.DeleteResponse{}
	err = delResp.FillResponse(resp)
	if err != nil {
		return resp.StatusCode, nil, err
	}

	if resp.StatusCode == http.StatusOK {
		return resp.StatusCode, delResp, nil
	} else {
		logger.ApiClientLogger.Println("[RESTClient] http.Delete StatusCode not http.StatusOK", err)
		return resp.StatusCode, delResp, errors.New("StatusCode not 200")
	}
}

func (c *RESTClient) WatchAll() (watch.Interface, error) {
	resourceURL := c.WatchURL()
	resp, err := http.Get(resourceURL)

	if err != nil {
		logger.ApiClientLogger.Println("[RESTClient] WatchAll Failed: ", err)
		// sleep some time before retry
		time.Sleep(time.Second * time.Duration(ReconnectInterval))
		return nil, err
	}

	decoder := watch.NewEtcdEventDecoder(resp.Body, c.resourceType)
	reporter := watch.NewDefaultReporter()
	streamWatcher := watch.NewStreamWatcher(decoder, reporter)

	return streamWatcher, nil
}

func (c *RESTClient) Watch(name string) (watch.Interface, error) {
	resourceURL := c.WatchURL() + name
	resp, err := http.Get(resourceURL)

	if err != nil {
		logger.ApiClientLogger.Printf("[RESTClient] Watch %v %v Failed: %v\n", c.resourceType, name, err)
		// sleep some time before retry
		time.Sleep(time.Second * time.Duration(ReconnectInterval))
		return nil, err
	}

	logger.ApiClientLogger.Printf("[RESTClient] Watch %v %v start\n", c.resourceType, name)

	decoder := watch.NewEtcdEventDecoder(resp.Body, c.resourceType)
	reporter := watch.NewDefaultReporter()
	streamWatcher := watch.NewStreamWatcher(decoder, reporter)

	return streamWatcher, nil
}

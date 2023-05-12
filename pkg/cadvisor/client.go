package cadvisor

import (
	"github.com/google/cadvisor/client"
	clientv2 "github.com/google/cadvisor/client/v2"
	info "github.com/google/cadvisor/info/v1"
	cadvisorapiv2 "github.com/google/cadvisor/info/v2"
	"log"
)

type Client struct {
	staticClient    *client.Client
	v2StaticClient  *clientv2.Client
	streamingClient *client.Client
}

// More info: https://github.com/google/cadvisor/tree/master/client

func NewClient(url string) Interface {

	nilCli := &Client{
		staticClient:    nil,
		v2StaticClient:  nil,
		streamingClient: nil,
	}

	staticClient, err := client.NewClient(url)
	if err != nil {
		log.Printf("[cadvisor] NewClient tried to make staticClient but got error %v\n", err)
		return nilCli
	}

	v2StaticClient, err := clientv2.NewClient(url)
	if err != nil {
		log.Printf("[cadvisor] NewClient tried to make v2StaticClient but got error %v\n", err)
		return nilCli
	}

	streamingClient, err := client.NewClient(url)
	if err != nil {
		log.Printf("[cadvisor] NewClient tried to make streamingClient but got error %v\n", err)
		return nilCli
	}

	return &Client{
		staticClient:    staticClient,
		v2StaticClient:  v2StaticClient,
		streamingClient: streamingClient,
	}
}

func (c *Client) Start() error {

	return nil
}

func (c *Client) DockerContainer(name string, req *info.ContainerInfoRequest) (info.ContainerInfo, error) {
	//TODO implement me
	panic("implement me")
}

// ContainerInfo Given a container name and a ContainerInfoRequest, will return all information about the specified container.
func (c *Client) ContainerInfo(name string, req *info.ContainerInfoRequest) (*info.ContainerInfo, error) {
	return c.staticClient.ContainerInfo(name, req)
}

func (c *Client) ContainerInfoV2(name string, options cadvisorapiv2.RequestOptions) (map[string]cadvisorapiv2.ContainerInfo, error) {
	return c.v2StaticClient.Stats(name, &options)
}

func (c *Client) SubcontainerInfo(name string, req *info.ContainerInfoRequest) (map[string]*info.ContainerInfo, error) {
	containers, err := c.staticClient.SubcontainersInfo(name, req)
	if err != nil && len(containers) == 0 {
		return nil, err
	}
	res := make(map[string]*info.ContainerInfo, len(containers))
	for _, container := range containers {
		res[container.Name] = &container
	}
	return res, err
}

// MachineInfo This method returns a cadvisor/v1.MachineInfo struct with all the fields filled in
func (c *Client) MachineInfo() (*info.MachineInfo, error) {
	return c.staticClient.MachineInfo()
}

func (c *Client) VersionInfo() (*info.VersionInfo, error) {
	attr, err := c.v2StaticClient.Attributes()
	if err != nil {
		return nil, err
	}
	cadvisorVersion, err := c.v2StaticClient.VersionInfo()
	if err != nil {
		return nil, err
	}
	return &info.VersionInfo{
		CadvisorVersion:    attr.CadvisorVersion,
		KernelVersion:      attr.KernelVersion,
		DockerAPIVersion:   attr.DockerAPIVersion,
		DockerVersion:      attr.DockerVersion,
		ContainerOsVersion: attr.ContainerOsVersion,
		CadvisorRevision:   cadvisorVersion,
	}, nil
}

func (c *Client) WatchAllEvents(containerName string, includeSubcontainers bool) (chan *info.Event, error) {

	// params := "?creation_events=true&stream=true&oom_events=true&deletion_events=true"
	params := "?all_events=true&stream=true"

	// if IncludeSubcontainers is false, only events occurring in the specific
	// container, and not the subcontainers, will be returned
	if includeSubcontainers {
		params += "&subcontainers=true"
	}

	// the absolute container name for which the event occurred
	url := containerName + params

	einfo := make(chan *info.Event)
	go func() {
		err := c.streamingClient.EventStreamingInfo(url, einfo)
		if err != nil {
			log.Printf("[cadvisor] got error retrieving event info: %v\n", err)
			return
		}
	}()
	return einfo, nil
}

func (c *Client) AllDockerContainers(query *info.ContainerInfoRequest) (cinfo []info.ContainerInfo, err error) {
	return c.staticClient.AllDockerContainers(query)
}

//func (c *Client) startStreamingClient(url string) {
//	einfo := make(chan *info.Event)
//	go func() {
//		err := c.streamingClient.EventStreamingInfo(url, einfo)
//		if err != nil {
//			log.Printf("[cadvisor] got error retrieving event info: %v\n", err)
//			return
//		}
//	}()
//	for ev := range einfo {
//		log.Printf("[cadvisor] streaming einfo: %v\n", ev)
//	}
//}

package cri

import (
	"context"
	"fmt"
	dt "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"io"
	"log"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/types"
)

func NewDocker() (Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &dockerClient{Client: cli}, nil
}

func (c *dockerClient) Close() {
	soundClose(c.Client)
}

type dockerClient struct {
	Client *client.Client
}

func (c *dockerClient) ContainerStart(ctx context.Context, name string) error {
	return c.Client.ContainerStart(ctx, c.ContainerId(ctx, name), dt.ContainerStartOptions{})
}

func (c *dockerClient) ContainerStatus(ctx context.Context, id string) (bool, int, error) {
	resp, err := c.Client.ContainerInspect(ctx, id)
	if err != nil {
		return false, 0, err
	}
	return resp.State.Running, resp.State.ExitCode, nil
}
func (c *dockerClient) ContainerIP(ctx context.Context, id string) (string, error) {
	resp, err := c.Client.ContainerInspect(ctx, id)
	if err != nil {
		return "", err
	}
	return resp.NetworkSettings.IPAddress, nil
}

func soundClose(cli *client.Client) {
	err := cli.Close()
	if err != nil {
		log.Println(err.Error())
	}
}

func (c *dockerClient) ContainerCreate(ctx context.Context, cnt core.Container) (string, error) {
	if len(cnt.Master) == 0 {
		return c.containerMasterCreate(ctx, cnt)
	} else {
		return c.containerSlaverCreate(ctx, cnt)
	}
}

func (c *dockerClient) ContainerRemove(ctx context.Context, name string) error {
	return c.Client.ContainerRemove(ctx, c.ContainerId(ctx, name), dt.ContainerRemoveOptions{
		RemoveVolumes: false,
		RemoveLinks:   false,
		Force:         true,
	})

}

func (c *dockerClient) containerMasterCreate(ctx context.Context, cnt core.Container) (string, error) {
	if cnt.Master != "" {
		return "", fmt.Errorf("HasMaster")
	}
	if err := c.handleImagePull(ctx, cnt); err != nil {
		return "", err
	}
	resp, err := c.Client.ContainerCreate(ctx, buildMasterContainerConfig(cnt), buildMasterHostConfig(cnt), nil, nil, cnt.Name)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (c *dockerClient) containerSlaverCreate(ctx context.Context, cnt core.Container) (string, error) {
	if cnt.Master == "" {
		return "", fmt.Errorf("NoMaster")
	}
	if err := c.handleImagePull(ctx, cnt); err != nil {
		return "", err
	}
	resp, err := c.Client.ContainerCreate(ctx, buildSlaverContainerConfig(cnt.Master, cnt), buildSlaverHostConfig(cnt.Master, cnt), nil, nil, cnt.Name)
	if err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (c *dockerClient) ContainerId(ctx context.Context, name string) string {
	list, err := c.Client.ContainerList(ctx, dt.ContainerListOptions{All: true})
	if err != nil {
		return ""
	}

	for _, c := range list {
		for _, n := range c.Names {
			if n == "/"+name {
				return c.ID
			}
		}
	}
	return ""
}

func buildMasterContainerConfig(cnt core.Container) *container.Config {
	return &container.Config{
		Tty:        cnt.TTY,
		StdinOnce:  cnt.StdinOnce,
		Cmd:        append(cnt.Command, cnt.Args...),
		Image:      cnt.Image,
		WorkingDir: cnt.WorkingDir,
	}
}

func buildSlaverContainerConfig(master string, cnt core.Container) *container.Config {
	return &container.Config{
		Hostname:        "",
		Domainname:      "",
		User:            "",
		AttachStdin:     false,
		AttachStdout:    false,
		AttachStderr:    false,
		ExposedPorts:    nil,
		Tty:             cnt.TTY,
		OpenStdin:       false,
		StdinOnce:       cnt.StdinOnce,
		Env:             buildEnv(cnt),
		Cmd:             append(cnt.Command, cnt.Args...),
		Healthcheck:     nil,
		ArgsEscaped:     false,
		Image:           cnt.Image,
		Volumes:         nil,
		WorkingDir:      cnt.WorkingDir,
		Entrypoint:      nil,
		NetworkDisabled: false,
		MacAddress:      "",
		OnBuild:         nil,
		Labels:          nil,
		StopSignal:      "",
		StopTimeout:     nil,
		Shell:           nil,
	}
}

func buildEnv(cnt core.Container) []string {
	var ret []string
	for _, ev := range cnt.Env {
		ret = append(ret, ev.Name+"="+ev.Value)
	}
	return ret
}

func buildMasterHostConfig(cnt core.Container) *container.HostConfig {
	return &container.HostConfig{
		Mounts: buildMount(cnt),
	}
}

func buildSlaverHostConfig(master string, cnt core.Container) *container.HostConfig {
	res := container.Resources{}

	for k, r := range cnt.Resources.Limits {
		switch k {
		case types.ResourceCPU:
			q, err := types.ParseQuantity(k, r)
			if err != nil {
				log.Fatalf("cpu quota not recognized %v", r)
			}
			res.NanoCPUs = int64(1_000_000_000 * q / 1024)
			break
		case types.ResourceMemory:
			q, err := types.ParseQuantity(k, r)
			if err != nil {
				log.Fatalf("memory quota not recognized %v", r)
			}
			res.Memory = int64(q * 1024 * 1204)
			break
		}
	}
	return &container.HostConfig{
		Binds:           nil,
		ContainerIDFile: "",
		LogConfig:       container.LogConfig{},
		NetworkMode:     container.NetworkMode("container:" + master),
		PortBindings:    nil,
		RestartPolicy:   container.RestartPolicy{},
		AutoRemove:      false,
		VolumeDriver:    "",
		VolumesFrom:     nil,
		CapAdd:          nil,
		CapDrop:         nil,
		CgroupnsMode:    "",
		DNS:             nil,
		DNSOptions:      nil,
		DNSSearch:       nil,
		ExtraHosts:      nil,
		GroupAdd:        nil,
		IpcMode:         "",
		Cgroup:          container.CgroupSpec("container:" + master),
		Links:           nil,
		OomScoreAdj:     0,
		PidMode:         "",
		Privileged:      false,
		PublishAllPorts: false,
		ReadonlyRootfs:  false,
		SecurityOpt:     nil,
		StorageOpt:      nil,
		Tmpfs:           nil,
		UTSMode:         "",
		UsernsMode:      "",
		ShmSize:         0,
		Sysctls:         nil,
		Runtime:         "",
		ConsoleSize:     [2]uint{},
		Isolation:       "",
		Resources:       res,
		Mounts:          buildMount(cnt),
		MaskedPaths:     nil,
		ReadonlyPaths:   nil,
		Init:            nil,
	}
}

func buildMount(cnt core.Container) []mount.Mount {
	mnt := make([]mount.Mount, 0)
	for _, m := range cnt.VolumeMounts {
		mnt = append(mnt, mount.Mount{
			Type:   mount.TypeVolume,
			Source: m.Name,
			Target: m.MountPath,
		})
	}

	return mnt
}

func (c *dockerClient) handleImagePull(ctx context.Context, cnt core.Container) error {
	switch cnt.ImagePullPolicy {
	default:
		fallthrough
	case core.PullIfNotPresent:
		//FIXME: check if present
		fallthrough
	case core.PullAlways:
		out, err := c.Client.ImagePull(ctx, cnt.Image, dt.ImagePullOptions{})
		if err != nil {
			return err
		}
		_, err = io.Copy(io.Discard, out)
		return err
	case core.PullNever:
		return nil
	}
}

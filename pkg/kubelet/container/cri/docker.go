package cri

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"log"
	"minik8s/pkg/api/core"
	"sync"
	"time"
)

func NewDocker() (CriClient, error) {
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

func soundClose(cli *client.Client) {
	err := cli.Close()
	if err != nil {
		log.Println(err.Error())
	}
}

func (c *dockerClient) ContainerEnsure(ctx context.Context, cnt core.Container, wg *sync.WaitGroup) {
	defer wg.Done()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer soundClose(cli)
	if err := handlImagePullPolicy(ctx, cli, cnt.Image, cnt.ImagePullPolicy); err != nil {
		log.Println(err.Error())
		return
	}

	id := c.containerGet(ctx, cnt.Name)

	if len(id) == 0 {
	create:
		for {
			select {
			case <-ctx.Done():
				return
			default:
				resp, err := cli.ContainerCreate(ctx, buildContainerConfig(cnt), buildHostConfig(cnt), nil, nil, cnt.Name)
				if err != nil {
					log.Println(err.Error())
					time.Sleep(1 * time.Second)
					continue
				} else {
					id = resp.ID

					break create
				}
			}
		}

	}

	if err := cli.ContainerStart(ctx, id, types.ContainerStartOptions{}); err != nil {
		log.Println(err.Error())
	}
	for {
		select {
		case <-ctx.Done():
			if err := cli.ContainerRemove(context.Background(), id, types.ContainerRemoveOptions{Force: true}); err != nil {
				log.Println(err.Error())
			}
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}

}

func (c *dockerClient) containerGet(ctx context.Context, name string) string {
	list, err := c.Client.ContainerList(ctx, types.ContainerListOptions{All: true})
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

func (c *dockerClient) VolumeCreate(ctx context.Context, name string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer soundClose(cli)
	if _, err := cli.VolumeCreate(ctx, volume.VolumeCreateBody{Name: name}); err != nil {
		return err
	}

	return nil
}

func (c *dockerClient) VolumeRemove(ctx context.Context, name string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer soundClose(cli)
	if err := cli.VolumeRemove(ctx, name, true); err != nil {
		return err
	}
	return nil
}

func handlImagePullPolicy(ctx context.Context, cli *client.Client, image string, policy core.PullPolicy) error {
	//FIXME policy is ignored and pull always is used
	_, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	return err
}

func buildContainerConfig(cnt core.Container) *container.Config {
	return &container.Config{
		Tty:        cnt.TTY,
		StdinOnce:  cnt.StdinOnce,
		Cmd:        append(cnt.Command, cnt.Args...),
		Image:      cnt.Image,
		WorkingDir: cnt.WorkingDir,
	}
}

func buildHostConfig(cnt core.Container) *container.HostConfig {
	return &container.HostConfig{
		Mounts: buildMount(cnt),
		RestartPolicy: container.RestartPolicy{
			Name:              "always",
			MaximumRetryCount: 0, //TODO
		},
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

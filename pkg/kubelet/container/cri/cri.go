package cri

import (
	"context"
	"minik8s/pkg/api/core"
)

type Client interface {
	//ContainerEnsure(ctx context.Context, cnt core.Container, group *sync.WaitGroup)
	//ContainerCleanEnsure(name string) error
	// VolumeCreate VolumeRemove FIXME they should be separated to csi, this is ugly
	//VolumeCreate(name string) error
	//VolumeRemove(name string) error
	ContainerCreate(ctx context.Context, cnt core.Container) (string, error)
	ContainerRemove(ctx context.Context, name string) error
	ContainerStart(ctx context.Context, name string) error
	ContainerInspect(ctx context.Context, name string) (bool, error)
	ContainerId(ctx context.Context, id string) string
	Close()
}

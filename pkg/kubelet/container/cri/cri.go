package cri

import (
	"context"
	"minik8s/pkg/api/core"
	"sync"
)

type CriClient interface {
	ContainerEnsure(ctx context.Context, cnt core.Container, group *sync.WaitGroup)
	//ContainerCleanEnsure(name string) error
	// VolumeCreate VolumeRemove FIXME they should be separated to csi, this is ugly
	//VolumeCreate(name string) error
	//VolumeRemove(name string) error
	Close()
}

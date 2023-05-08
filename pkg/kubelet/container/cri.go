package container

import (
	"context"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/opencontainers/runtime-spec/specs-go"
	"log"
	"minik8s/pkg/api/core"
)

type CriClient interface {
	SetAddress(address string)
	SetNamespace(namespace string)
	CreateContainer(container core.Container) error
	CleanContainer(name string) error
}

func NewCriClient() CriClient {
	return &criClient{
		address:   "/run/containerd/containerd.sock",
		namespace: "kubelet",
	}
}

type criClient struct {
	address   string
	namespace string
}

func (c *criClient) SetAddress(addr string) {
	c.address = addr
}

func (c *criClient) SetNamespace(ns string) {
	c.namespace = ns
}

func (c *criClient) CreateContainer(cnt core.Container) error {
	pp := log.Prefix()
	defer log.SetPrefix(pp)
	log.SetPrefix("[create]")
	log.Println(cnt.Name)
	ctx := namespaces.WithNamespace(context.Background(), c.namespace)
	clnt, err := containerd.New(c.address)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer func(clnt *containerd.Client) {
		err := clnt.Close()
		if err != nil {
			log.Println(err.Error())
		}
	}(clnt)
	ncos, err := buildNCOpts(ctx, clnt, cnt)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	cont, err := clnt.NewContainer(ctx, cnt.Name, ncos...)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	task, err := cont.NewTask(ctx, cio.NullIO)
	//task, err := cont.NewTask(ctx, cio.NewCreator(cio.WithStdio)) // This is for test
	if err != nil {
		log.Println(err.Error())
		return err
	}
	if err := task.Start(ctx); err != nil {
		return err
	}
	return nil
}

func buildNCOpts(ctx context.Context, clnt *containerd.Client, cnt core.Container) ([]containerd.NewContainerOpts, error) {
	img, err := buildImage(ctx, clnt, cnt)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return []containerd.NewContainerOpts{buildSnapshotNCOpt(cnt, img), buildSpecNCOpt(cnt, img)}, nil
}

func buildSnapshotNCOpt(cnt core.Container, img containerd.Image) containerd.NewContainerOpts {
	return containerd.WithNewSnapshot(cnt.Name, img)
}

func buildImage(ctx context.Context, clnt *containerd.Client, cnt core.Container) (containerd.Image, error) {
	//switch cnt.ImagePullPolicy {
	//case core.PullAlways:
	//	return clnt.Pull(ctx, cnt.Image, containerd.WithPullUnpack)
	//case core.PullNever:
	//	imgs, err := clnt.ListImages(ctx, cnt.Image)
	//	if err != nil {
	//		return nil, err
	//	}
	//	if len(imgs) == 0 {
	//		return nil, fmt.Errorf("image %s not exists", cnt.Image)
	//	}
	//	return imgs[0], nil
	//default:
	//	fallthrough
	//case core.PullIfNotPresent:
	//	imgs, err := clnt.ListImages(ctx, cnt.Image)
	//	if err != nil {
	//		return nil, err
	//	}
	//	if len(imgs) == 0 {
	//		return clnt.Pull(ctx, cnt.Image)
	//	}
	//	return imgs[0], nil
	//}
	// FIXME the pull policy is ignored
	return clnt.Pull(ctx, cnt.Image, containerd.WithPullUnpack)
}

func buildSpecNCOpt(cnt core.Container, img oci.Image) containerd.NewContainerOpts {
	return containerd.WithNewSpec(buildSOpts(cnt, img)...)
}

func buildSOpts(cnt core.Container, img oci.Image) []oci.SpecOpts {
	opts := []oci.SpecOpts{oci.WithImageConfig(img)}
	if carg := buildCommandArgSOpt(cnt); carg != nil {
		opts = append(opts, carg)
	}
	if wd := buildWorkingDirSOpt(cnt); wd != nil {
		opts = append(opts, wd)
	}
	if env := buildEnvSOpt(cnt); env != nil {
		opts = append(opts, env)
	}
	if tty := buildTtySOpt(cnt); tty != nil {
		opts = append(opts, tty)
	}
	if m := buildMountSOpt(cnt); m != nil {
		opts = append(opts, m)
	}
	return opts
}

func buildCommandArgSOpt(cnt core.Container) oci.SpecOpts {
	var args []string
	if cnt.Command == nil {
		return nil
	}
	args = append(args, cnt.Command...)
	if cnt.Args != nil {
		args = append(args, cnt.Args...)
	}
	return oci.WithProcessArgs(args...)
}

func buildWorkingDirSOpt(cnt core.Container) oci.SpecOpts {
	if len(cnt.WorkingDir) != 0 {
		return oci.WithProcessCwd(cnt.WorkingDir)
	}
	return nil
}

func buildEnvSOpt(cnt core.Container) oci.SpecOpts {
	var envs []string
	for _, e := range cnt.Env {
		envs = append(envs, e.Name+"="+e.Value)
	}
	if len(cnt.Env) != 0 {
		return oci.WithEnv(envs)
	}
	return nil
}

//TODO Resources

func buildMountSOpt(cnt core.Container) oci.SpecOpts {

	if cnt.VolumeMounts == nil {
		return nil
	}

	ms := []specs.Mount{}

	for _, m := range cnt.VolumeMounts {
		ms = append(ms,
			specs.Mount{
				Destination: m.MountPath,
				Type:        "bind",
				Source:      m.Name,
				Options:     []string{"bind"}, //in linux, mount bind can mount a folder to another folder
			})
	}
	return oci.WithMounts(ms)
}

//TODO Stdin
//TODO StdinOnce

func buildTtySOpt(cnt core.Container) oci.SpecOpts {
	if cnt.TTY {
		return oci.WithTTY
	}
	return nil
}

func (c *criClient) CleanContainer(name string) error {
	pp := log.Prefix()
	defer log.SetPrefix(pp)
	log.SetPrefix("[clean]")
	log.Println(name)
	clnt, err := containerd.New(c.address)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	defer func(clnt *containerd.Client) {
		err := clnt.Close()
		if err != nil {
			log.Println(err.Error())
		}
	}(clnt)
	ctx := namespaces.WithNamespace(context.Background(), c.namespace)

	if err := cleanContainer(ctx, clnt, name); err != nil {
		log.Println(err.Error())
	}

	if err := cleanSnapshot(ctx, clnt, name); err != nil {
		log.Println(err.Error())
	}
	return nil
}

func cleanContainer(ctx context.Context, clnt *containerd.Client, name string) error {
	cnt, err := clnt.LoadContainer(ctx, name)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	if ex, err := cleanTask(ctx, cnt); err != nil {
		log.Println(err.Error())
	} else {
		log.Println(ex.Result())
	}

	if err := cnt.Delete(ctx); err != nil {
		log.Println(err.Error())
	}

	return nil
}

func cleanTask(ctx context.Context, cnt containerd.Container) (*containerd.ExitStatus, error) {
	task, err := cnt.Task(ctx, nil)
	if err != nil {
		if err.Error() == "NotFound" {
			return nil, nil
		}
		return nil, err
	}
	s, err := task.Status(ctx)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	if s.Status != containerd.Stopped {
		return task.Delete(ctx, containerd.WithProcessKill)
	} else {
		return task.Delete(ctx)
	}
}

func cleanSnapshot(ctx context.Context, clnt *containerd.Client, name string) error {
	return clnt.SnapshotService("").Remove(ctx, name)
}

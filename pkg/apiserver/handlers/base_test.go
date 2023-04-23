package handlers

import (
	"github.com/gin-gonic/gin"
	"minik8s/pkg/api/core"
	"testing"
)

func TestHandleDeleteNode(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleDeleteNode(tt.args.c)
		})
	}
}

func TestHandleDeleteNodes(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleDeleteNodes(tt.args.c)
		})
	}
}

func TestHandleDeletePod(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleDeletePod(tt.args.c)
		})
	}
}

func TestHandleDeleteService(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleDeleteService(tt.args.c)
		})
	}
}

func TestHandleGetNode(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleGetNode(tt.args.c)
		})
	}
}

func TestHandleGetNodeStatus(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleGetNodeStatus(tt.args.c)
		})
	}
}

func TestHandleGetNodes(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleGetNodes(tt.args.c)
		})
	}
}

func TestHandleGetPod(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleGetPod(tt.args.c)
		})
	}
}

func TestHandleGetPodStatus(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleGetPodStatus(tt.args.c)
		})
	}
}

func TestHandleGetPods(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleGetPods(tt.args.c)
		})
	}
}

func TestHandleGetService(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleGetService(tt.args.c)
		})
	}
}

func TestHandleGetServiceStatus(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleGetServiceStatus(tt.args.c)
		})
	}
}

func TestHandleGetServices(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleGetServices(tt.args.c)
		})
	}
}

func TestHandlePostNode(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandlePostNode(tt.args.c)
		})
	}
}

func TestHandlePostPod(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandlePostPod(tt.args.c)
		})
	}
}

func TestHandlePostService(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandlePostService(tt.args.c)
		})
	}
}

func TestHandlePutNode(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandlePutNode(tt.args.c)
		})
	}
}

func TestHandlePutNodeStatus(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandlePutNodeStatus(tt.args.c)
		})
	}
}

func TestHandlePutPod(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandlePutPod(tt.args.c)
		})
	}
}

func TestHandlePutPodStatus(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandlePutPodStatus(tt.args.c)
		})
	}
}

func TestHandlePutService(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandlePutService(tt.args.c)
		})
	}
}

func TestHandlePutServiceStatus(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandlePutServiceStatus(tt.args.c)
		})
	}
}

func TestHandleWatchNode(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleWatchNode(tt.args.c)
		})
	}
}

func TestHandleWatchNodes(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleWatchNodes(tt.args.c)
		})
	}
}

func TestHandleWatchPod(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleWatchPod(tt.args.c)
		})
	}
}

func TestHandleWatchPods(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleWatchPods(tt.args.c)
		})
	}
}

func TestHandleWatchService(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleWatchService(tt.args.c)
		})
	}
}

func TestHandleWatchServices(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HandleWatchServices(tt.args.c)
		})
	}
}

func Test_handleDeleteObject(t *testing.T) {
	type args struct {
		c  *gin.Context
		ty core.ApiObjectType
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleDeleteObject(tt.args.c, tt.args.ty)
		})
	}
}

func Test_handleGetObject(t *testing.T) {
	type args struct {
		c  *gin.Context
		ty core.ApiObjectType
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleGetObject(tt.args.c, tt.args.ty)
		})
	}
}

func Test_handleGetObjectStatus(t *testing.T) {
	type args struct {
		c           *gin.Context
		ty          core.ApiObjectType
		resourceURL string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleGetObjectStatus(tt.args.c, tt.args.ty, tt.args.resourceURL)
		})
	}
}

func Test_handleGetObjects(t *testing.T) {
	type args struct {
		c  *gin.Context
		ty core.ApiObjectType
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleGetObjects(tt.args.c, tt.args.ty)
		})
	}
}

func Test_handlePostObject(t *testing.T) {
	type args struct {
		c  *gin.Context
		ty core.ApiObjectType
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlePostObject(tt.args.c, tt.args.ty)
		})
	}
}

func Test_handlePutObject(t *testing.T) {
	type args struct {
		c  *gin.Context
		ty core.ApiObjectType
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlePutObject(tt.args.c, tt.args.ty)
		})
	}
}

func Test_handlePutObjectStatus(t *testing.T) {
	type args struct {
		c       *gin.Context
		ty      core.ApiObjectType
		etcdURL string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlePutObjectStatus(tt.args.c, tt.args.ty, tt.args.etcdURL)
		})
	}
}

func Test_handleWatchObjectAndStatus(t *testing.T) {
	type args struct {
		c           *gin.Context
		ty          core.ApiObjectType
		resourceURL string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleWatchObjectAndStatus(tt.args.c, tt.args.ty, tt.args.resourceURL)
		})
	}
}

func Test_handleWatchObjectsAndStatus(t *testing.T) {
	type args struct {
		c           *gin.Context
		ty          core.ApiObjectType
		resourceURL string
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleWatchObjectsAndStatus(tt.args.c, tt.args.ty, tt.args.resourceURL)
		})
	}
}

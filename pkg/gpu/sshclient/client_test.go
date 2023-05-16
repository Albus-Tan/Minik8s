package sshclient

import (
	"context"
	"log"
	"minik8s/utils"
	"os"
	"path/filepath"
	"testing"
)

func Test_sshClient_SubmitJob(t *testing.T) {
	cli := New()
	ctx, cancel := context.WithCancel(context.Background())
	cli.Run(ctx)
	defer cancel()

	path, _ := os.Getwd()
	path1 := filepath.Join(path, "../cuda/sum_matrix/sum_matrix.cu")
	path2 := filepath.Join(path, "../cuda/sum_matrix/sum_matrix.slurm")
	id, err := cli.SubmitCudaJob(utils.GenerateUID(), path1, path2, "sum_matrix")
	if err != nil {
		return
	}
	log.Printf("Job id %s\n", id)

}

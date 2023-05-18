package jobclient

import (
	"context"
	"log"
	"minik8s/utils"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func Test_jobClient_SubmitJob(t *testing.T) {
	cli := New()
	ctx, cancel := context.WithCancel(context.Background())
	cli.Run(ctx)
	defer cancel()

	path, _ := os.Getwd()
	path = filepath.Join(path, "../cuda/sum_matrix/sum_matrix.cu")
	content := `#!/bin/bash
#SBATCH --job-name=Test_jobClient_SubmitJob
#SBATCH --partition=dgx2
#SBATCH --output=sum_matrix.out
#SBATCH --error=sum_matrix.err
#SBATCH -N 1
#SBATCH --ntasks-per-node=1
#SBATCH --cpus-per-task=1
#SBATCH --gres=gpu:1

ulimit -s unlimited
ulimit -l unlimited

module load gcc/8.3.0 cuda/10.1.243-gcc-8.3.0

./sum_matrix
`
	id, err := cli.SubmitCudaJob(utils.GenerateUID(), path, content, "sum_matrix")
	if err != nil {
		return
	}
	log.Printf("Job id %s\n", id)
	isFinished, _ := cli.CheckJobFinish(id)
	for !isFinished {
		isFinished, _ = cli.CheckJobFinish(id)
		log.Printf("Job id %s not finished\n", id)
		time.Sleep(time.Second * 3)
	}
	log.Printf("Job id %s finished\n", id)
}

//func Test_jobClient_DownloadResult(t *testing.T) {
//
//	var jobUID = "57ec0192-cd65-4d69-9cfa-399683e8e9b8"
//
//	cli := New()
//	ctx, cancel := context.WithCancel(context.Background())
//	cli.Run(ctx)
//	defer cancel()
//
//	path, _ := os.Getwd()
//	path = filepath.Join(path, "../cuda/sum_matrix")
//
//	result, err := cli.DownloadResult(jobUID, path, "sum_matrix")
//	if err != nil {
//		log.Printf("err: %s\n", err)
//		return
//	}
//
//	if result {
//		log.Printf("DownloadResult success\n")
//	} else {
//		log.Printf("DownloadResult failed\n")
//	}
//}

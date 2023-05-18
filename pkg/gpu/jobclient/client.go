package jobclient

import (
	"context"
	"errors"
	"fmt"
	"github.com/melbahja/goph"
	"github.com/pkg/sftp"
	"log"
	"minik8s/config"
	"minik8s/pkg/api/core"
	"path/filepath"
	"strings"
	"time"
)

type Interface interface {
	Run(ctx context.Context)
	SubmitCudaJob(jobUID string, cuFilePath string, slurmFilePath string, objectFileName string) (jobId string, err error)
	CheckJobFinish(jobId string) (bool, error)
	GetJobState(jobId string) (string, error)
	DownloadResult(jobUID string, localFilePath string, resultFileName string) (bool, error)
}

type jobClient struct {
	client *goph.Client
	sftpc  *sftp.Client
}

func New() Interface {
	return &jobClient{
		client: nil,
		sftpc:  nil,
	}
}

func (c *jobClient) CheckJobFinish(jobId string) (bool, error) {
	cmd := fmt.Sprintf("sacct -j %s | tail -n +3 | awk '{print $1, $2, $3, $4, $5, $6, $7}'", jobId)
	res, err := c.executeCommand(cmd)
	if err != nil {
		log.Printf("[jobClient] CheckJobFinish executeCommand err: %v\n", err)
		return false, err
	}

	// TODO: modify CheckJobFinish bash
	resp := string(res)
	log.Printf("[jobClient] resp: %v\n", resp)
	rows := strings.Split(resp, "\n")
	if len(rows) > 0 {
		row := rows[0]
		cols := strings.Split(row, " ")
		if len(cols) == 7 {
			if cols[5] == "COMPLETED" {
				return true, nil
			} else {
				return false, nil
			}
		}
	}
	return false, errors.New(fmt.Sprintf("Job %s not found\n", jobId))
}

func (c *jobClient) GetJobState(jobId string) (string, error) {
	cmd := fmt.Sprintf("sacct -j %s | tail -n +3 | awk '{print $1, $2, $3, $4, $5, $6, $7}'", jobId)
	res, err := c.executeCommand(cmd)
	if err != nil {
		log.Printf("[jobClient] CheckJobFinish executeCommand err: %v\n", err)
		return "", err
	}

	// TODO: modify CheckJobFinish bash
	resp := string(res)
	log.Printf("[jobClient] resp: %v\n", resp)
	rows := strings.Split(resp, "\n")
	if len(rows) > 0 {
		row := rows[0]
		cols := strings.Split(row, " ")
		if len(cols) == 7 {
			return cols[5], nil
		}
	}
	return string(core.JobMissing), errors.New(fmt.Sprintf("Job %s not found\n", jobId))
}

func (c *jobClient) DownloadResult(jobUID string, localFilePath string, resultFileName string) (bool, error) {

	dirName := config.HPCJobDirPrefix + jobUID
	fullDirName := config.HPCHomeDir + dirName
	resultFileOutputName := resultFileName + config.OutputFileSuffix
	resultFileErrorName := resultFileName + config.ErrorFileSuffix
	resultFileOutputDstPath := filepath.ToSlash(filepath.Join(fullDirName, resultFileOutputName))
	resultFileErrorDstPath := filepath.ToSlash(filepath.Join(fullDirName, resultFileErrorName))

	err := c.download(filepath.Join(localFilePath, resultFileOutputName), resultFileOutputDstPath)
	if err != nil {
		return false, err
	}
	err = c.download(filepath.Join(localFilePath, resultFileErrorName), resultFileErrorDstPath)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *jobClient) SubmitCudaJob(jobUID string, cuFilePath string, slurmFilePath string, objectFileName string) (jobId string, err error) {

	dirName := config.HPCJobDirPrefix + jobUID
	fullDirName := config.HPCHomeDir + dirName
	cuFileName := filepath.Base(cuFilePath)
	slurmFileName := filepath.Base(slurmFilePath)
	cuFileDstPath := filepath.ToSlash(filepath.Join(fullDirName, cuFileName))
	slurmFileDstPath := filepath.ToSlash(filepath.Join(fullDirName, slurmFileName))
	objectFileDstPath := filepath.ToSlash(filepath.Join(fullDirName, objectFileName))

	res, err := c.executeCommand("mkdir " + dirName)
	if err != nil {
		log.Printf("[jobClient] SubmitCudaJob executeCommand err: %v\n", err)
		return "-1", err
	}

	err = c.upload(cuFilePath, cuFileDstPath)
	if err != nil {
		log.Printf("[jobClient] SubmitCudaJob upload file %v to %v err: %v\n", cuFilePath, cuFileDstPath, err)
		return "-1", err
	}

	err = c.upload(slurmFilePath, slurmFileDstPath)
	if err != nil {
		log.Printf("[jobClient] SubmitCudaJob upload file %v to %v err: %v\n", slurmFilePath, slurmFileDstPath, err)
		return "-1", err
	}

	cmd := fmt.Sprintf("module load gcc/8.3.0 cuda/10.1.243-gcc-8.3.0 && nvcc %s -o %s -lcublas && sbatch %s", cuFileDstPath, objectFileDstPath, slurmFileDstPath)
	res, err = c.executeCommand(cmd)
	if err != nil {
		log.Printf("[jobClient] SubmitCudaJob executeCommandsAndGetLastOutput err: %v\n", err)
		return "-1", err
	}

	var jobID string
	n, err := fmt.Sscanf(string(res), "Submitted batch job %s", &jobID)
	if err != nil || n != 1 {
		return "-1", err
	}
	return jobID, nil
}

func (c *jobClient) Run(ctx context.Context) {
	syncChan := make(chan bool)
	go c.run(ctx, syncChan)
	<-syncChan
}

func (c *jobClient) run(ctx context.Context, syncChan chan bool) {

	if c.client == nil {
		var err error
		c.client, err = goph.New(config.HPCUsername, config.PiHost, goph.Password(config.HPCPassword))
		if err != nil {
			log.Fatal(fmt.Sprintf("[jobClient] New ssh client failed: %v\n", err))
		}
		c.newSftp()
	}

	// Defer closing the network connection.
	defer c.client.Close()

	log.Printf("[jobClient] ssh client connect success\n")

	// send signal through syncChan to tell Run client init finish
	syncChan <- true

	<-ctx.Done()
}

func (c *jobClient) newSftp() {
	var err error
	if c.sftpc == nil {
		c.sftpc, err = c.client.NewSftp()
		if err != nil {
			log.Fatal(fmt.Sprintf("[jobClient] New Sftp client failed: %v\n", err))
		}
	}
}

func (c *jobClient) upload(localFilePath string, remoteFilePath string) error {
	return c.client.Upload(localFilePath, remoteFilePath)
}

func (c *jobClient) download(localFilePath string, remoteFilePath string) error {
	return c.client.Download(remoteFilePath, localFilePath)
}

func (c *jobClient) executeCommand(cmd string) ([]byte, error) {
	return c.client.Run(cmd)
}

func (c *jobClient) executeCommandsAndGetLastOutput(cmds []string) (out []byte, err error) {
	for _, cmd := range cmds {
		out, err = c.executeCommand(cmd)
		if err != nil {
			return nil, err
		}
		time.Sleep(time.Second)
	}
	return out, nil
}

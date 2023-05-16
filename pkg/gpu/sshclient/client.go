package sshclient

import (
	"context"
	"fmt"
	"github.com/melbahja/goph"
	"github.com/pkg/sftp"
	"log"
	"minik8s/config"
	"path/filepath"
)

type Interface interface {
	Run(ctx context.Context)
	SubmitCudaJob(name string, cuFilePath string, slurmFilePath string, objectFileName string) (jobId string, err error)
}

type sshClient struct {
	client *goph.Client
	sftpc  *sftp.Client
}

func New() Interface {
	return &sshClient{
		client: nil,
		sftpc:  nil,
	}
}

func (c *sshClient) SubmitCudaJob(name string, cuFilePath string, slurmFilePath string, objectFileName string) (jobId string, err error) {

	dirName := config.HPCJobDirPrefix + name
	fullDirName := config.HPCHomeDir + dirName
	cuFileName := filepath.Base(cuFilePath)
	slurmFileName := filepath.Base(slurmFilePath)
	cuFileDstPath := filepath.ToSlash(filepath.Join(fullDirName, cuFileName))
	slurmFileDstPath := filepath.ToSlash(filepath.Join(fullDirName, slurmFileName))
	objectFileDstPath := filepath.ToSlash(filepath.Join(fullDirName, objectFileName))

	res, err := c.executeCommand("mkdir " + dirName)
	if err != nil {
		log.Printf("[sshClient] SubmitCudaJob executeCommandsAndGetLastOutput err: %v\n", err)
		return "-1", err
	}

	err = c.upload(cuFilePath, cuFileDstPath)
	if err != nil {
		log.Printf("[sshClient] SubmitCudaJob upload file %v to %v err: %v\n", cuFilePath, cuFileDstPath, err)
		return "-1", err
	}

	err = c.upload(slurmFilePath, slurmFileDstPath)
	if err != nil {
		log.Printf("[sshClient] SubmitCudaJob upload file %v to %v err: %v\n", slurmFilePath, slurmFileDstPath, err)
		return "-1", err
	}

	cmds := []string{
		fmt.Sprintf("module load gcc/8.3.0 cuda/10.1.243-gcc-8.3.0 && nvcc %s -o %s -lcublas", cuFileDstPath, objectFileDstPath),
		fmt.Sprintf("sbatch %s", slurmFileDstPath),
	}
	res, err = c.executeCommandsAndGetLastOutput(cmds)
	if err != nil {
		log.Printf("[sshClient] SubmitCudaJob executeCommandsAndGetLastOutput err: %v\n", err)
		return "-1", err
	}

	var jobID string
	n, err := fmt.Sscanf(string(res), "Submitted batch job %s", &jobID)
	if err != nil || n != 1 {
		return "-1", err
	}
	return jobID, nil
}

func (c *sshClient) Run(ctx context.Context) {
	syncChan := make(chan bool)
	go c.run(ctx, syncChan)
	<-syncChan
}

func (c *sshClient) run(ctx context.Context, syncChan chan bool) {

	if c.client == nil {
		var err error
		c.client, err = goph.New(config.HPCUsername, config.PiHost, goph.Password(config.HPCPassword))
		if err != nil {
			log.Fatal(fmt.Sprintf("[sshClient] New ssh client failed: %v\n", err))
		}
		c.newSftp()
	}

	// Defer closing the network connection.
	defer c.client.Close()

	log.Printf("[sshClient] ssh client connect success\n")

	// send signal through syncChan to tell Run client init finish
	syncChan <- true

	<-ctx.Done()
}

func (c *sshClient) newSftp() {
	var err error
	if c.sftpc == nil {
		c.sftpc, err = c.client.NewSftp()
		if err != nil {
			log.Fatal(fmt.Sprintf("[sshClient] New Sftp client failed: %v\n", err))
		}
	}
}

func (c *sshClient) upload(localFilePath string, remoteFilePath string) error {
	return c.client.Upload(localFilePath, remoteFilePath)
}

func (c *sshClient) download(localFilePath string, remoteFilePath string) error {
	return c.client.Download(remoteFilePath, localFilePath)
}

func (c *sshClient) executeCommand(cmd string) ([]byte, error) {
	return c.client.Run(cmd)
}

func (c *sshClient) executeCommandsAndGetLastOutput(cmds []string) (out []byte, err error) {
	for _, cmd := range cmds {
		out, err = c.executeCommand(cmd)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

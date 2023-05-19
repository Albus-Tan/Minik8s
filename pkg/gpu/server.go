package gpu

import (
	"context"
	"errors"
	"fmt"
	"minik8s/config"
	"minik8s/pkg/api/core"
	"minik8s/pkg/api/types"
	"minik8s/pkg/api/watch"
	"minik8s/pkg/apiclient"
	client "minik8s/pkg/apiclient/interface"
	"minik8s/pkg/apiclient/listwatch"
	"minik8s/pkg/gpu/jobclient"
	"minik8s/pkg/logger"
	"minik8s/utils/datastructure"
	"net/http"
	"time"
)

type Server interface {
	Run(ctx context.Context, cancel context.CancelFunc)
}

func NewServer() Server {

	jobClient, _ := apiclient.NewRESTClient(types.JobObjectType)
	jobListWatcher := listwatch.NewListWatchFromClient(jobClient)

	return &server{
		cli:            jobclient.New(),
		jobClient:      jobClient,
		jobListWatcher: jobListWatcher,
		jobQueue:       datastructure.NewConcurrentQueue(),
	}
}

type server struct {
	cli            jobclient.Interface
	jobClient      client.Interface
	jobListWatcher listwatch.ListerWatcher
	jobQueue       datastructure.IConcurrentQueue
}

func (s *server) Run(ctx context.Context, cancel context.CancelFunc) {

	logger.GpuServerLogger.Printf("[GpuServer] start\n")
	defer logger.GpuServerLogger.Printf("[GpuServer] init finish\n")

	s.cli.Run(ctx)

	// run watch jobs
	go func() {
		defer cancel()
		err := s.listAndWatchJobs(ctx.Done())
		if err != nil {
			logger.GpuServerLogger.Printf("[GpuServer] listAndWatchJobs failed, err: %v\n", err)
		}
	}()

	go func() {
		defer cancel()
		s.runJobWorker(ctx)
	}()

	go func() {
		defer cancel()
		s.periodicallyCheckJobState()
	}()

}

const jobStateCheckInterval = 30 * time.Second

func (s *server) periodicallyCheckJobState() {
	for {
		logger.GpuServerLogger.Printf("[periodicallyCheckJobState] check start\n")

		time.Sleep(jobStateCheckInterval)
		jobList, err := s.jobListWatcher.List()
		if err != nil {
			logger.GpuServerLogger.Printf("[periodicallyCheckJobState] jobListWatcher list failed\n")
			continue
		}

		logger.GpuServerLogger.Printf("[periodicallyCheckJobState] jobList %v\n", jobList)

		jobs := jobList.GetIApiObjectArr()
		logger.GpuServerLogger.Printf("[periodicallyCheckJobState] %v jobs to check in apiserver storage\n", len(jobs))
		for i, item := range jobs {
			job := item.(*core.Job)
			jobID := job.Status.JobID
			if jobID == "" {
				logger.GpuServerLogger.Printf("[periodicallyCheckJobState] job %v do not have JobID now\n", i)
				continue
			}

			if core.JobFinished(job.Status.State) {
				logger.GpuServerLogger.Printf("[periodicallyCheckJobState] job %v, JobID %v has already done, state %v\n", i, jobID, job.Status.State)
				continue
			}

			jobState, err := s.cli.GetJobState(jobID)
			logger.GpuServerLogger.Printf("[periodicallyCheckJobState] job %v, JobID %v, state %v\n", i, jobID, jobState)
			if jobState == "" || jobState == string(core.JobMissing) || jobState == string(job.Status.State) {
				logger.GpuServerLogger.Printf("[periodicallyCheckJobState] job %v state not found or unchange\n", i)
				continue
			}

			job.Status.State = core.JobState(jobState)

			// send binding result to apiserver
			code, _, err := s.jobClient.Put(job.UID, job)
			if err != nil {
				for code == http.StatusConflict {
					jobItem, _ := s.jobClient.Get(job.UID)
					job = jobItem.(*core.Job)

					// modify job content
					jobState, err = s.cli.GetJobState(jobID)
					job.Status.State = core.JobState(jobState)

					code, _, err = s.jobClient.Put(job.UID, job)
				}
			}

			logger.GpuServerLogger.Printf("[periodicallyCheckJobState] job %v state update\n", i)
		}

		logger.GpuServerLogger.Printf("[periodicallyCheckJobState] jobs state check finish\n")
	}
}

const defaultWorkerSleepInterval = time.Duration(3) * time.Second

func (s *server) runJobWorker(ctx context.Context) {

	// go wait.UntilWithContext(ctx, rsc.worker, time.Second)
	for {
		select {
		case <-ctx.Done():
			logger.GpuServerLogger.Printf("[worker] ctx.Done() received, worker of GpuServer exit\n")
			return
		default:
			for s.processNextJob() {
			}
			time.Sleep(defaultWorkerSleepInterval)
		}
	}

}

func (s *server) processNextJob() bool {

	job := s.dequeueJob()
	if job == nil {
		return false
	}

	// submit job
	jobId, err := s.submitJob(job)
	if err != nil {
		logger.GpuServerLogger.Printf("[processNextJob] submit job uid %v error: %v\n", job.UID, err)
		return false
	}

	// modify job content
	job.Status.JobID = jobId
	jobState, err := s.cli.GetJobState(jobId)
	job.Status.State = core.JobState(jobState)

	// send binding result to apiserver
	code, _, err := s.jobClient.Put(job.UID, job)
	if err != nil {
		for code == http.StatusConflict {
			jobItem, _ := s.jobClient.Get(job.UID)
			job = jobItem.(*core.Job)

			// modify job content
			job.Status.JobID = jobId
			jobState, err = s.cli.GetJobState(jobId)
			job.Status.State = core.JobState(jobState)

			code, _, err = s.jobClient.Put(job.UID, job)
		}
		return code == http.StatusOK
	}

	logger.GpuServerLogger.Printf("[processNextJob] submit job uid %v finish\n", job.UID)

	return true
}

var (
	errorStopRequested = errors.New("stop requested")
)

func (s *server) listAndWatchJobs(stopCh <-chan struct{}) error {

	// list all jobs and push into jobsQueue
	jobsList, err := s.jobListWatcher.List()
	if err != nil {
		return err
	}

	jobItems := jobsList.GetIApiObjectArr()
	for _, item := range jobItems {
		job := item.(*core.Job)
		s.enqueueJob(job)
	}

	// start watch jobs change
	var w watch.Interface
	w, err = s.jobListWatcher.Watch()
	if err != nil {
		return err
	}

	err = s.handleWatchJobs(w, stopCh)
	w.Stop() // stop watch

	if err == errorStopRequested {
		return nil
	}

	return err

}

func (s *server) enqueueJob(job *core.Job) {
	s.jobQueue.Enqueue(job)
	logger.GpuServerLogger.Printf("[enqueueJob] job %v enqueued\n", job.UID)
}

func (s *server) dequeueJob() *core.Job {
	jobItem, exist := s.jobQueue.Dequeue()
	if exist {
		j := jobItem.(*core.Job)
		logger.GpuServerLogger.Printf("[dequeueJob] job %v equeued\n", j.UID)
		return j
	} else {
		logger.GpuServerLogger.Printf("[dequeueJob] queue empty\n")
		return nil
	}
}

func (s *server) handleWatchJobs(w watch.Interface, stopCh <-chan struct{}) error {
	eventCount := 0
loop:
	for {
		select {
		case <-stopCh:
			return errorStopRequested
		case event, ok := <-w.ResultChan():
			if !ok {
				break loop
			}
			logger.GpuServerLogger.Printf("[handleWatchJobs] event %v\n", event)
			logger.GpuServerLogger.Printf("[handleWatchJobs] event object %v\n", event.Object)
			eventCount += 1

			switch event.Type {
			case watch.Added:
				newJob := (event.Object).(*core.Job)
				s.enqueueJob(newJob)
				logger.GpuServerLogger.Printf("[handleWatchJobs] new Job event, handle job %v created\n", newJob.UID)
			case watch.Modified:
				newJob := (event.Object).(*core.Job)
				go s.handleJobModified(newJob)
			case watch.Deleted:
				// ignore
			case watch.Bookmark:
				panic("[handleWatchJobs] watchHandler Event Type watch.Bookmark received")
			case watch.Error:
				panic("[handleWatchJobs] watchHandler Event Type watch.Error received")
			default:
				panic("[handleWatchJobs] watchHandler Unknown Event Type received")
			}
		}
	}
	return nil
}

func (s *server) submitJob(job *core.Job) (jobId string, err error) {
	slurmFile := GenerateJobScript(job)
	jobId, err = s.cli.SubmitCudaJob(job.UID, job.Spec.CuFilePath, slurmFile, job.Spec.ResultFileName)
	return jobId, err
}

func (s *server) handleJobModified(job *core.Job) {
	// if job finish successfully
	if job.Status.State == core.JobCompleted {
		logger.GpuServerLogger.Printf("[handleJobModified] handling Job COMPLETED, jobId %v\n", job.Status.JobID)

		downloaded, _ := s.downloadJobResult(job)

		// download failed, change job state back to RUNNING for future check state
		// to observe finish again
		if !downloaded {
			job.Status.State = core.JobRunning
			// send binding result to apiserver
			code, _, err := s.jobClient.Put(job.UID, job)
			if err != nil {
				for code == http.StatusConflict {
					jobItem, _ := s.jobClient.Get(job.UID)
					job = jobItem.(*core.Job)
					job.Status.State = core.JobRunning
					code, _, err = s.jobClient.Put(job.UID, job)
				}
				return
			}
			return
		}

		// job finished and download result success
		// TODO: notify user
		logger.GpuServerLogger.Printf("[handleJobModified] handling Job COMPLETED, jobID %v, result downloaded success\n", job.Status.JobID)

	} else if job.Status.State == core.JobFailed {
		logger.GpuServerLogger.Printf("[handleJobModified] handling Job FAILED, jobId %v\n", job.Status.JobID)
		// job finished and failed
		// TODO: notify user

	}
}

const retryDownloadTimes = 10

func (s *server) downloadJobResult(job *core.Job) (downloaded bool, err error) {
	downloaded = false
	retry := 0
	for !downloaded && retry < retryDownloadTimes {
		downloaded, err = s.cli.DownloadResult(job.UID, job.Spec.ResultFilePath, job.Spec.ResultFileName)
		if err != nil {
			logger.GpuServerLogger.Printf("[handleJobModified] downloadJobResult for jobId %v failed: %v\n", job.Status.JobID, err)
		}
		retry += 1
	}
	return downloaded, err
}

func GenerateJobScript(job *core.Job) string {

	mailRemindTemplate := `#SBATCH --mail-type=%s
#SBATCH --mail-user=%s%s
`
	mailRemind := ""
	if job.Spec.Args.Mail != nil {
		mailRemind = fmt.Sprintf(
			mailRemindTemplate,
			job.Spec.Args.Mail.Type,
			job.Spec.Args.Mail.UserName,
			config.MailAddressSuffix,
		)
	}

	template := `#!/bin/bash
#SBATCH --job-name=%s
#SBATCH --partition=dgx2
#SBATCH --output=%s.out
#SBATCH --error=%s.err
#SBATCH -N 1
#SBATCH --ntasks-per-node=%d
#SBATCH --cpus-per-task=%d
#SBATCH --gres=gpu:%d
%s

ulimit -s unlimited
ulimit -l unlimited

module load gcc/8.3.0 cuda/10.1.243-gcc-8.3.0

./%s
`
	numTasksPerNode := 0
	if job.Spec.Args.NumTasksPerNode != 0 {
		numTasksPerNode = job.Spec.Args.NumTasksPerNode
	} else {
		numTasksPerNode = 1
	}
	cpusPerTask := 0
	if job.Spec.Args.CpusPerTask != 0 {
		cpusPerTask = job.Spec.Args.CpusPerTask
	} else {
		cpusPerTask = 1
	}
	gpuResources := 0
	if job.Spec.Args.GpuResources != 0 {
		gpuResources = job.Spec.Args.GpuResources
	} else {
		gpuResources = 1
	}

	script := fmt.Sprintf(
		template,
		job.UID,
		job.Spec.ResultFileName,
		job.Spec.ResultFileName,
		numTasksPerNode,
		cpusPerTask,
		gpuResources,
		mailRemind,
		job.Spec.ResultFileName,
	)
	return script
}

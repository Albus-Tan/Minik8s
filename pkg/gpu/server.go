package gpu

import (
	"context"
	"errors"
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
				// ignore
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
	jobId, err = s.cli.SubmitCudaJob(job.UID, job.Spec.CuFilePath, job.Spec.SlurmFilePath, job.Spec.ObjectFileName)
	return jobId, err
}

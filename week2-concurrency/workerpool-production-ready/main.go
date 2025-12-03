package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// represents a unit of work
type Job struct {
	ID   int
	Data string
}

// represents the output of processing a job
type Result struct {
	Job   Job
	Error error
	Value string
}

// manages a pool of workers
type WorkerPool struct {
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	numWorkers int
	jobs       chan Job
	results    chan Result
}

// creates a new Worker Pool
func NewWorkerPool(numWorkers, jobQueueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	return &WorkerPool{
		ctx:        ctx,
		cancel:     cancel,
		numWorkers: numWorkers,
		jobs:       make(chan Job, jobQueueSize),
		results:    make(chan Result, jobQueueSize),
	}
}

// calls the worker method which are the workers of the worker pool using a go routine
func (wp *WorkerPool) Start() {
	fmt.Printf("Starting %d Workers...\n", wp.numWorkers)

	for workerId := 1; workerId <= wp.numWorkers; workerId++ {
		wp.wg.Add(1)
		go wp.Worker(workerId)
	}
}

func (wp *WorkerPool) Worker(workerId int) {
	defer wp.wg.Done()

	fmt.Printf("Worker %d started\n", workerId)

	for {
		select {
		case job, ok := <-wp.jobs:
			if !ok {
				fmt.Printf("Worker %d: jobs channel closed\n", workerId)
				return
			}

			fmt.Printf("Worker %d processing job %d\n", workerId, job.ID)
			result := wp.ProcessJob(job)

			select {
			case wp.results <- result:
			case <-wp.ctx.Done():
				fmt.Printf("Worker %d: context cancelled\n", workerId)
			}
		case <-wp.ctx.Done():
			fmt.Printf("Worker %d: shutting down\n", workerId)
			return
		}
	}
}

func (wp *WorkerPool) ProcessJob(job Job) Result {
	// simulate work
	time.Sleep(500 * time.Millisecond)

	return Result{
		Job:   job,
		Value: fmt.Sprintf("Processed: %s", job.Data),
		Error: nil,
	}
}

func (wp *WorkerPool) Submit(job Job) error {
	select {
	case wp.jobs <- job:
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("Worker pool is shutting down")
	}
}

func (wp *WorkerPool) Results() <-chan Result {
	return wp.results
}

func (wp *WorkerPool) Stop() {
	fmt.Println("Stopping Worker Pool...")
	wp.cancel()       // send the cancel signal to the rest of the workers, so that the workers wont send to result channel anymore as it is close and would cause panic
	close(wp.jobs)    // close jobs channel, no more jobs should be sent to the channel
	wp.wg.Wait()      // wait for workers already processing jobs to finish
	close(wp.results) // close results after worker with already processing hobs finish work
	fmt.Println("Worker Pool Stopped")
}

func main() {
	fmt.Println("========= Production Ready Worker Pool =========")

	numworkers := 3 // no. of workers
	numJobs := 10   // no. of Jobs
	// create the worker pool
	pool := NewWorkerPool(numworkers, numJobs)

	// start the worker pool
	pool.Start()

	go func() {
		for i := 1; i <= numJobs; i++ {
			job := Job{
				ID:   i,
				Data: fmt.Sprintf("Task-%d", i),
			}

			if err := pool.Submit(job); err != nil {
				fmt.Printf("Failed to submit job: %v\n", err)
				return
			}
		}

		time.Sleep(2 * time.Second)
		pool.Stop()
	}()

	// Collect results
	for results := range pool.Results() {
		if results.Error != nil {
			fmt.Printf("Job %d failed %v\n", results.Job.ID, results.Error)
		} else {
			fmt.Printf("Job %d completed: %s\n", results.Job.ID, results.Value)
		}
	}

	fmt.Println("All jobs processed")
}

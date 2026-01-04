package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Pipeline manages the data ingestion flow.
type Pipeline struct {
	jobChan    chan Job
	results    chan jobResultPair
	processor  Processor
	sinks      []Sink
	workerCount int
	metrics    *Metrics
	wg         sync.WaitGroup
	quit       chan struct{}
}

type jobResultPair struct {
	Job    Job
	Result Result
}

// NewPipeline creates a new ingestion pipeline.
func NewPipeline(workerCount int, bufferSize int, processor Processor, sinks ...Sink) *Pipeline {
	return &Pipeline{
		jobChan:    make(chan Job, bufferSize),
		results:    make(chan jobResultPair, bufferSize),
		processor:  processor,
		sinks:      sinks,
		workerCount: workerCount,
		metrics:    NewMetrics(),
		quit:       make(chan struct{}),
	}
}

// Start initializes the workers and starts processing.
func (p *Pipeline) Start(ctx context.Context) {
	zap.S().Info("Starting Ingestion Pipeline", "workers", p.workerCount)

	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}

	// Result handler
	p.wg.Add(1)
	go p.resultHandler(ctx)
}

// Stop gracefully shuts down the pipeline.
func (p *Pipeline) Stop() {
	close(p.quit)
	close(p.jobChan) // Stop accepting new jobs
	p.wg.Wait()
	zap.S().Info("Pipeline stopped")
}

// Submit adds a job to the pipeline.
func (p *Pipeline) Submit(job Job) error {
	select {
	case p.jobChan <- job:
		p.metrics.JobSubmitted()
		return nil
	default:
		p.metrics.JobDropped()
		return fmt.Errorf("pipeline buffer full, job dropped")
	}
}

// GetJobChan returns the channel for submitting jobs.
func (p *Pipeline) GetJobChan() chan<- Job {
	return p.jobChan
}

func (p *Pipeline) worker(ctx context.Context, id int) {
	defer p.wg.Done()
	for {
		select {
		case <-p.quit:
			return
		case <-ctx.Done():
			return
		case job, ok := <-p.jobChan:
			if !ok {
				return
			}
			p.processJob(ctx, job)
		}
	}
}

func (p *Pipeline) processJob(ctx context.Context, job Job) {
	start := time.Now()
	zap.S().Debug("Processing job", "job_id", job.ID, "source", job.Source)

	// Validation
	if err := ValidateJob(job); err != nil {
		zap.S().Error("Job validation failed", "job_id", job.ID, "error", err)
		p.results <- jobResultPair{Job: job, Result: Result{JobID: job.ID, Success: false, Error: err, ProcessedAt: time.Now()}}
		return
	}

	// Processing
	res, err := p.processor.Process(ctx, job)
	if err != nil {
		zap.S().Error("Processing failed", "job_id", job.ID, "error", err)
		// Retry Logic could go here (re-queueing)
		if job.RetryCount < job.MaxRetries {
			job.RetryCount++
			zap.S().Info("Retrying job", "job_id", job.ID, "attempt", job.RetryCount)
			// Non-blocking re-queue
			go p.Submit(job)
			return
		}
		res = Result{JobID: job.ID, Success: false, Error: err, ProcessedAt: time.Now()}
	}

	p.metrics.JobProcessed(time.Since(start))
	p.results <- jobResultPair{Job: job, Result: res}
}

func (p *Pipeline) resultHandler(ctx context.Context) {
	defer p.wg.Done()
	for {
		select {
		case <-p.quit:
			return
		case <-ctx.Done():
			return
		case pair, ok := <-p.results:
			if !ok {
				return
			}
			if pair.Result.Success {
				for _, sink := range p.sinks {
					if err := sink.Save(ctx, pair.Job, pair.Result); err != nil {
						zap.S().Error("Sink save failed", "job_id", pair.Job.ID, "error", err)
					}
				}
			}
		}
	}
}

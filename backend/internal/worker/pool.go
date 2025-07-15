package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type JobType string

const (
	JobTypeAnalyzeURL JobType = "analyze_url"
	JobTypeCrawlURL   JobType = "crawl_url"
	JobTypeCleanup    JobType = "cleanup"
)

type Job struct {
	ID        string
	Type      JobType
	Payload   interface{}
	Retry     int
	MaxRetry  int
	CreatedAt time.Time
}

type JobResult struct {
	Job   Job
	Error error
	Data  interface{}
}

type JobHandler func(ctx context.Context, job Job) (interface{}, error)

type WorkerPool struct {
	workerCount int
	jobQueue    chan Job
	resultQueue chan JobResult
	quit        chan struct{}
	wg          sync.WaitGroup
	handlers    map[JobType]JobHandler
	logger      *slog.Logger
}

func NewWorkerPool(workerCount int, queueSize int, logger *slog.Logger) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		jobQueue:    make(chan Job, queueSize),
		resultQueue: make(chan JobResult, queueSize),
		quit:        make(chan struct{}),
		handlers:    make(map[JobType]JobHandler),
		logger:      logger,
	}
}

func (wp *WorkerPool) RegisterHandler(jobType JobType, handler JobHandler) {
	wp.handlers[jobType] = handler
}

func (wp *WorkerPool) Start(ctx context.Context) {
	wp.logger.Info("Starting worker pool",
		slog.Int("workers", wp.workerCount),
		slog.Int("queue_size", cap(wp.jobQueue)))

	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}

	go wp.processResults(ctx)
}

func (wp *WorkerPool) Stop() {
	wp.logger.Info("Stopping worker pool")
	close(wp.quit)
	wp.wg.Wait()
	close(wp.jobQueue)
	close(wp.resultQueue)
	wp.logger.Info("Worker pool stopped")
}

func (wp *WorkerPool) AddJob(job Job) error {
	select {
	case wp.jobQueue <- job:
		wp.logger.Debug("Job added to queue",
			slog.String("job_id", job.ID),
			slog.String("job_type", string(job.Type)))
		return nil
	default:
		return fmt.Errorf("job queue is full")
	}
}

func (wp *WorkerPool) AddJobWithTimeout(job Job, timeout time.Duration) error {
	select {
	case wp.jobQueue <- job:
		wp.logger.Debug("Job added to queue",
			slog.String("job_id", job.ID),
			slog.String("job_type", string(job.Type)))
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("timeout adding job to queue")
	}
}

func (wp *WorkerPool) worker(ctx context.Context, workerID int) {
	defer wp.wg.Done()

	wp.logger.Debug("Worker started", slog.Int("worker_id", workerID))

	for {
		select {
		case job := <-wp.jobQueue:
			wp.processJob(ctx, workerID, job)
		case <-wp.quit:
			wp.logger.Debug("Worker stopping", slog.Int("worker_id", workerID))
			return
		case <-ctx.Done():
			wp.logger.Debug("Worker context cancelled", slog.Int("worker_id", workerID))
			return
		}
	}
}

func (wp *WorkerPool) processJob(ctx context.Context, workerID int, job Job) {
	start := time.Now()

	wp.logger.Debug("Processing job",
		slog.Int("worker_id", workerID),
		slog.String("job_id", job.ID),
		slog.String("job_type", string(job.Type)))

	handler, exists := wp.handlers[job.Type]
	if !exists {
		wp.logger.Error("No handler registered for job type",
			slog.String("job_type", string(job.Type)),
			slog.String("job_id", job.ID))

		wp.resultQueue <- JobResult{
			Job:   job,
			Error: fmt.Errorf("no handler registered for job type: %s", job.Type),
		}
		return
	}

	jobCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	data, err := handler(jobCtx, job)

	duration := time.Since(start)

	if err != nil {
		wp.logger.Error("Job processing failed",
			slog.String("job_id", job.ID),
			slog.String("job_type", string(job.Type)),
			slog.String("error", err.Error()),
			slog.Duration("duration", duration))

		if job.Retry < job.MaxRetry {
			job.Retry++
			wp.logger.Info("Retrying job",
				slog.String("job_id", job.ID),
				slog.Int("retry", job.Retry),
				slog.Int("max_retry", job.MaxRetry))

			time.Sleep(time.Duration(job.Retry) * time.Second)
			wp.AddJob(job)
			return
		}
	} else {
		wp.logger.Debug("Job processed successfully",
			slog.String("job_id", job.ID),
			slog.String("job_type", string(job.Type)),
			slog.Duration("duration", duration))
	}

	select {
	case wp.resultQueue <- JobResult{Job: job, Error: err, Data: data}:
	default:
		wp.logger.Warn("Result queue full, dropping result",
			slog.String("job_id", job.ID))
	}
}

func (wp *WorkerPool) processResults(ctx context.Context) {
	for {
		select {
		case result := <-wp.resultQueue:
			wp.handleJobResult(result)
		case <-ctx.Done():
			return
		}
	}
}

func (wp *WorkerPool) handleJobResult(result JobResult) {
	if result.Error != nil {
		wp.logger.Error("Job failed",
			slog.String("job_id", result.Job.ID),
			slog.String("job_type", string(result.Job.Type)),
			slog.String("error", result.Error.Error()))
	} else {
		wp.logger.Debug("Job completed successfully",
			slog.String("job_id", result.Job.ID),
			slog.String("job_type", string(result.Job.Type)))
	}
}

func (wp *WorkerPool) GetStats() PoolStats {
	return PoolStats{
		WorkerCount:    wp.workerCount,
		JobsInQueue:    len(wp.jobQueue),
		ResultsInQueue: len(wp.resultQueue),
		QueueCapacity:  cap(wp.jobQueue),
	}
}

type PoolStats struct {
	WorkerCount    int `json:"worker_count"`
	JobsInQueue    int `json:"jobs_in_queue"`
	ResultsInQueue int `json:"results_in_queue"`
	QueueCapacity  int `json:"queue_capacity"`
}

type PriorityWorkerPool struct {
	*WorkerPool
	highPriorityQueue chan Job
	lowPriorityQueue  chan Job
}

func NewPriorityWorkerPool(workerCount int, queueSize int, logger *slog.Logger) *PriorityWorkerPool {
	wp := NewWorkerPool(workerCount, queueSize, logger)

	return &PriorityWorkerPool{
		WorkerPool:        wp,
		highPriorityQueue: make(chan Job, queueSize/2),
		lowPriorityQueue:  make(chan Job, queueSize/2),
	}
}

func (pwp *PriorityWorkerPool) AddHighPriorityJob(job Job) error {
	select {
	case pwp.highPriorityQueue <- job:
		pwp.logger.Debug("High priority job added",
			slog.String("job_id", job.ID),
			slog.String("job_type", string(job.Type)))
		return nil
	default:
		return fmt.Errorf("high priority queue is full")
	}
}

func (pwp *PriorityWorkerPool) AddLowPriorityJob(job Job) error {
	select {
	case pwp.lowPriorityQueue <- job:
		pwp.logger.Debug("Low priority job added",
			slog.String("job_id", job.ID),
			slog.String("job_type", string(job.Type)))
		return nil
	default:
		return fmt.Errorf("low priority queue is full")
	}
}

type BatchJob struct {
	ID       string
	Jobs     []Job
	Callback func(results []JobResult)
}

type BatchWorkerPool struct {
	*WorkerPool
	batchSize int
	batch     []Job
	mu        sync.Mutex
}

func NewBatchWorkerPool(workerCount int, queueSize int, batchSize int, logger *slog.Logger) *BatchWorkerPool {
	wp := NewWorkerPool(workerCount, queueSize, logger)

	return &BatchWorkerPool{
		WorkerPool: wp,
		batchSize:  batchSize,
		batch:      make([]Job, 0, batchSize),
	}
}

func (bwp *BatchWorkerPool) AddJobToBatch(job Job) error {
	bwp.mu.Lock()
	defer bwp.mu.Unlock()

	bwp.batch = append(bwp.batch, job)

	if len(bwp.batch) >= bwp.batchSize {
		return bwp.processBatch()
	}

	return nil
}

func (bwp *BatchWorkerPool) processBatch() error {
	if len(bwp.batch) == 0 {
		return nil
	}

	batchJob := Job{
		ID:        fmt.Sprintf("batch_%d", time.Now().Unix()),
		Type:      "batch",
		Payload:   bwp.batch,
		CreatedAt: time.Now(),
	}

	err := bwp.AddJob(batchJob)
	if err != nil {
		return err
	}

	bwp.batch = make([]Job, 0, bwp.batchSize)
	return nil
}

func (bwp *BatchWorkerPool) FlushBatch() error {
	bwp.mu.Lock()
	defer bwp.mu.Unlock()

	return bwp.processBatch()
}
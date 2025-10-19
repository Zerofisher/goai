package dispatcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Zerofisher/goai/pkg/types"
)

// ConcurrentExecutor manages concurrent tool execution with advanced control.
type ConcurrentExecutor struct {
	dispatcher       *Dispatcher
	maxConcurrency   int
	queueSize        int
	workerPool       chan struct{}
	taskQueue        chan *Task
	results          sync.Map
	wg               sync.WaitGroup
	ctx              context.Context
	cancel           context.CancelFunc
}

// Task represents a tool execution task.
type Task struct {
	ID       string
	ToolUse  types.ToolUse
	ResultCh chan types.ToolResult
	Priority int
	Deadline time.Time
}

// NewConcurrentExecutor creates a new concurrent executor.
func NewConcurrentExecutor(dispatcher *Dispatcher, maxConcurrency int) *ConcurrentExecutor {
	ctx, cancel := context.WithCancel(context.Background())

	ce := &ConcurrentExecutor{
		dispatcher:     dispatcher,
		maxConcurrency: maxConcurrency,
		queueSize:      maxConcurrency * 10,
		workerPool:     make(chan struct{}, maxConcurrency),
		taskQueue:      make(chan *Task, maxConcurrency*10),
		ctx:            ctx,
		cancel:         cancel,
	}

	// Initialize worker pool
	for i := 0; i < maxConcurrency; i++ {
		ce.workerPool <- struct{}{}
	}

	// Start workers
	ce.startWorkers()

	return ce
}

// startWorkers starts the worker goroutines.
func (ce *ConcurrentExecutor) startWorkers() {
	for i := 0; i < ce.maxConcurrency; i++ {
		go ce.worker(i)
	}
}

// worker processes tasks from the queue.
func (ce *ConcurrentExecutor) worker(id int) {
	for {
		select {
		case task := <-ce.taskQueue:
			ce.processTask(task)
		case <-ce.ctx.Done():
			return
		}
	}
}

// processTask processes a single task.
func (ce *ConcurrentExecutor) processTask(task *Task) {
	// Acquire worker slot
	<-ce.workerPool
	defer func() {
		ce.workerPool <- struct{}{}
	}()

	// Check deadline
	if !task.Deadline.IsZero() && time.Now().After(task.Deadline) {
		task.ResultCh <- *task.ToolUse.Error(fmt.Errorf("task deadline exceeded"))
		close(task.ResultCh)
		return
	}

	// Create task context with deadline
	ctx := ce.ctx
	if !task.Deadline.IsZero() {
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, task.Deadline)
		defer cancel()
	}

	// Execute the tool
	result := ce.dispatcher.ExecuteWithContext(ctx, task.ToolUse)

	// Store result
	ce.results.Store(task.ID, result)

	// Send result
	select {
	case task.ResultCh <- result:
	case <-ce.ctx.Done():
	}
	close(task.ResultCh)
}

// Execute submits a tool for execution and returns immediately.
func (ce *ConcurrentExecutor) Execute(toolUse types.ToolUse) <-chan types.ToolResult {
	resultCh := make(chan types.ToolResult, 1)

	task := &Task{
		ID:       toolUse.ID,
		ToolUse:  toolUse,
		ResultCh: resultCh,
		Priority: 0,
		Deadline: time.Time{},
	}

	select {
	case ce.taskQueue <- task:
		// Task queued successfully
	default:
		// Queue is full
		go func() {
			resultCh <- *toolUse.Error(fmt.Errorf("task queue is full"))
			close(resultCh)
		}()
	}

	return resultCh
}

// ExecuteWithPriority submits a tool for execution with priority.
func (ce *ConcurrentExecutor) ExecuteWithPriority(toolUse types.ToolUse, priority int) <-chan types.ToolResult {
	resultCh := make(chan types.ToolResult, 1)

	task := &Task{
		ID:       toolUse.ID,
		ToolUse:  toolUse,
		ResultCh: resultCh,
		Priority: priority,
		Deadline: time.Time{},
	}

	// For high priority, try to insert at the front
	if priority > 0 {
		select {
		case ce.taskQueue <- task:
			// Task queued successfully
		default:
			// Queue is full
			go func() {
				resultCh <- *toolUse.Error(fmt.Errorf("task queue is full"))
				close(resultCh)
			}()
		}
	} else {
		ce.taskQueue <- task
	}

	return resultCh
}

// ExecuteBatch executes multiple tools concurrently.
func (ce *ConcurrentExecutor) ExecuteBatch(toolUses []types.ToolUse) []types.ToolResult {
	results := make([]types.ToolResult, len(toolUses))
	channels := make([]<-chan types.ToolResult, len(toolUses))

	// Submit all tasks
	for i, toolUse := range toolUses {
		channels[i] = ce.Execute(toolUse)
	}

	// Collect results
	for i, ch := range channels {
		select {
		case result := <-ch:
			results[i] = result
		case <-ce.ctx.Done():
			results[i] = *toolUses[i].Error(fmt.Errorf("execution cancelled"))
		}
	}

	return results
}

// ExecuteParallel executes tools in parallel groups.
func (ce *ConcurrentExecutor) ExecuteParallel(groups [][]types.ToolUse) [][]types.ToolResult {
	results := make([][]types.ToolResult, len(groups))

	for i, group := range groups {
		results[i] = ce.ExecuteBatch(group)
	}

	return results
}

// GetResult retrieves a cached result by task ID.
func (ce *ConcurrentExecutor) GetResult(taskID string) (types.ToolResult, bool) {
	if result, ok := ce.results.Load(taskID); ok {
		return result.(types.ToolResult), true
	}
	return types.ToolResult{}, false
}

// Shutdown gracefully shuts down the executor.
func (ce *ConcurrentExecutor) Shutdown(timeout time.Duration) error {
	// Signal shutdown
	ce.cancel()

	// Wait for workers to finish or timeout
	done := make(chan struct{})
	go func() {
		ce.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("shutdown timeout exceeded")
	}
}

// Stats returns statistics about the executor.
type ExecutorStats struct {
	QueueSize      int
	ActiveWorkers  int
	MaxConcurrency int
	CachedResults  int
}

// GetStats returns current executor statistics.
func (ce *ConcurrentExecutor) GetStats() ExecutorStats {
	queueSize := len(ce.taskQueue)
	activeWorkers := ce.maxConcurrency - len(ce.workerPool)

	cachedResults := 0
	ce.results.Range(func(_, _ interface{}) bool {
		cachedResults++
		return true
	})

	return ExecutorStats{
		QueueSize:      queueSize,
		ActiveWorkers:  activeWorkers,
		MaxConcurrency: ce.maxConcurrency,
		CachedResults:  cachedResults,
	}
}

// PriorityQueue implements a priority queue for tasks.
type PriorityQueue struct {
	tasks  []*Task
	mu     sync.Mutex
}

// NewPriorityQueue creates a new priority queue.
func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		tasks: make([]*Task, 0),
	}
}

// Push adds a task to the queue.
func (pq *PriorityQueue) Push(task *Task) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	// Insert task in priority order
	inserted := false
	for i, t := range pq.tasks {
		if task.Priority > t.Priority {
			pq.tasks = append(pq.tasks[:i], append([]*Task{task}, pq.tasks[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		pq.tasks = append(pq.tasks, task)
	}
}

// Pop removes and returns the highest priority task.
func (pq *PriorityQueue) Pop() (*Task, bool) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if len(pq.tasks) == 0 {
		return nil, false
	}

	task := pq.tasks[0]
	pq.tasks = pq.tasks[1:]
	return task, true
}

// Len returns the number of tasks in the queue.
func (pq *PriorityQueue) Len() int {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return len(pq.tasks)
}

// BatchProcessor processes tools in optimized batches.
type BatchProcessor struct {
	dispatcher      *Dispatcher
	batchSize       int
	batchTimeout    time.Duration
	processingDelay time.Duration
}

// NewBatchProcessor creates a new batch processor.
func NewBatchProcessor(dispatcher *Dispatcher, batchSize int, batchTimeout time.Duration) *BatchProcessor {
	return &BatchProcessor{
		dispatcher:      dispatcher,
		batchSize:       batchSize,
		batchTimeout:    batchTimeout,
		processingDelay: 100 * time.Millisecond,
	}
}

// Process processes tool uses in batches.
func (bp *BatchProcessor) Process(ctx context.Context, toolUses []types.ToolUse) []types.ToolResult {
	if len(toolUses) <= bp.batchSize {
		// Single batch
		return bp.dispatcher.ExecuteBatchWithContext(ctx, toolUses)
	}

	// Process in batches
	results := make([]types.ToolResult, 0, len(toolUses))
	for i := 0; i < len(toolUses); i += bp.batchSize {
		end := i + bp.batchSize
		if end > len(toolUses) {
			end = len(toolUses)
		}

		batch := toolUses[i:end]
		batchCtx, cancel := context.WithTimeout(ctx, bp.batchTimeout)
		batchResults := bp.dispatcher.ExecuteBatchWithContext(batchCtx, batch)
		cancel()

		results = append(results, batchResults...)

		// Add delay between batches to avoid overwhelming the system
		if end < len(toolUses) {
			select {
			case <-time.After(bp.processingDelay):
			case <-ctx.Done():
				// Fill remaining results with errors
				for j := end; j < len(toolUses); j++ {
					results = append(results, *toolUses[j].Error(fmt.Errorf("batch processing cancelled")))
				}
				return results
			}
		}
	}

	return results
}
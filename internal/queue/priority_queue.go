package queue

import (
	"container/heap"
	"sync"

	"github.com/vhvcorp/go-notification-service/internal/domain"
)

// Priority represents the priority level of an email job
type Priority int

const (
	// PriorityHigh for critical emails (alerts, security)
	PriorityHigh Priority = iota
	// PriorityNormal for regular transactional emails
	PriorityNormal
	// PriorityLow for marketing emails
	PriorityLow
)

// EmailJob represents an email job in the queue
type EmailJob struct {
	ID       string
	Priority Priority
	Request  *domain.SendEmailRequest
	Index    int // Index in the heap
}

// emailJobHeap implements heap.Interface
type emailJobHeap []*EmailJob

func (h emailJobHeap) Len() int { return len(h) }

func (h emailJobHeap) Less(i, j int) bool {
	// Lower priority value = higher priority (processed first)
	return h[i].Priority < h[j].Priority
}

func (h emailJobHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index = i
	h[j].Index = j
}

func (h *emailJobHeap) Push(x interface{}) {
	n := len(*h)
	job := x.(*EmailJob)
	job.Index = n
	*h = append(*h, job)
}

func (h *emailJobHeap) Pop() interface{} {
	old := *h
	n := len(old)
	job := old[n-1]
	old[n-1] = nil // Avoid memory leak
	job.Index = -1
	*h = old[0 : n-1]
	return job
}

// PriorityQueue is a thread-safe priority queue for email jobs
type PriorityQueue struct {
	jobs emailJobHeap
	mu   sync.Mutex
	cond *sync.Cond
}

// NewPriorityQueue creates a new priority queue
func NewPriorityQueue() *PriorityQueue {
	pq := &PriorityQueue{
		jobs: make(emailJobHeap, 0),
	}
	pq.cond = sync.NewCond(&pq.mu)
	heap.Init(&pq.jobs)
	return pq
}

// Push adds a job to the queue
func (pq *PriorityQueue) Push(job *EmailJob) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	heap.Push(&pq.jobs, job)
	pq.cond.Signal() // Wake up a waiting worker
}

// Pop removes and returns the highest priority job
// Blocks if the queue is empty
func (pq *PriorityQueue) Pop() *EmailJob {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	// Wait while queue is empty
	for pq.jobs.Len() == 0 {
		pq.cond.Wait()
	}

	job := heap.Pop(&pq.jobs).(*EmailJob)
	return job
}

// TryPop tries to pop a job without blocking
// Returns nil if queue is empty
func (pq *PriorityQueue) TryPop() *EmailJob {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	if pq.jobs.Len() == 0 {
		return nil
	}

	job := heap.Pop(&pq.jobs).(*EmailJob)
	return job
}

// Len returns the number of jobs in the queue
func (pq *PriorityQueue) Len() int {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	return pq.jobs.Len()
}

// IsEmpty returns true if the queue is empty
func (pq *PriorityQueue) IsEmpty() bool {
	return pq.Len() == 0
}

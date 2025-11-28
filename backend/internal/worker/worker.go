package worker

import (
	"context"
)

// Worker defines the interface for background workers
type Worker interface {
	// Start begins the worker's background processing
	Start(ctx context.Context)

	// Stop gracefully stops the worker
	Stop()

	// Name returns the worker's name for logging
	Name() string
}

// Manager handles multiple workers
type Manager struct {
	workers []Worker
}

// NewManager creates a new worker manager
func NewManager() *Manager {
	return &Manager{
		workers: make([]Worker, 0),
	}
}

// Register adds a worker to the manager
func (m *Manager) Register(w Worker) {
	m.workers = append(m.workers, w)
}

// StartAll starts all registered workers
func (m *Manager) StartAll(ctx context.Context) {
	for _, w := range m.workers {
		go w.Start(ctx)
	}
}

// StopAll stops all registered workers
func (m *Manager) StopAll() {
	for _, w := range m.workers {
		w.Stop()
	}
}

package terraform

import (
	"sync"
	"time"
)

// OperationLog stores logs for a terraform operation
type OperationLog struct {
	EnvironmentName string
	Operation       string // "create" or "destroy"
	Lines           []LogLine
	StartTime       time.Time
	EndTime         *time.Time
	Status          string // "running", "completed", "failed"
	Error           string
	mutex           sync.RWMutex
}

// LogLine represents a single log line
type LogLine struct {
	Timestamp time.Time
	Content   string
}

// LogStore manages operation logs
type LogStore struct {
	logs  map[string]*OperationLog
	mutex sync.RWMutex
}

var globalLogStore = &LogStore{
	logs: make(map[string]*OperationLog),
}

// GetLogStore returns the global log store
func GetLogStore() *LogStore {
	return globalLogStore
}

// CreateOperation creates a new operation log
func (s *LogStore) CreateOperation(envName, operation string) *OperationLog {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log := &OperationLog{
		EnvironmentName: envName,
		Operation:       operation,
		Lines:           []LogLine{},
		StartTime:       time.Now(),
		Status:          "running",
	}

	s.logs[envName] = log
	return log
}

// GetOperation retrieves an operation log
func (s *LogStore) GetOperation(envName string) *OperationLog {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.logs[envName]
}

// ListOperations returns all operation logs
func (s *LogStore) ListOperations() []*OperationLog {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	ops := make([]*OperationLog, 0, len(s.logs))
	for _, log := range s.logs {
		ops = append(ops, log)
	}
	return ops
}

// DeleteOperation removes an operation log
func (s *LogStore) DeleteOperation(envName string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.logs, envName)
}

// AddLine adds a log line to the operation
func (o *OperationLog) AddLine(content string) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.Lines = append(o.Lines, LogLine{
		Timestamp: time.Now(),
		Content:   content,
	})
}

// SetCompleted marks the operation as completed
func (o *OperationLog) SetCompleted() {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	now := time.Now()
	o.EndTime = &now
	o.Status = "completed"
}

// SetFailed marks the operation as failed
func (o *OperationLog) SetFailed(err error) {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	now := time.Now()
	o.EndTime = &now
	o.Status = "failed"
	if err != nil {
		o.Error = err.Error()
	}
}

// GetLines returns all log lines (thread-safe)
func (o *OperationLog) GetLines() []LogLine {
	o.mutex.RLock()
	defer o.mutex.RUnlock()

	lines := make([]LogLine, len(o.Lines))
	copy(lines, o.Lines)
	return lines
}

// GetStatus returns the current status (thread-safe)
func (o *OperationLog) GetStatus() string {
	o.mutex.RLock()
	defer o.mutex.RUnlock()
	return o.Status
}

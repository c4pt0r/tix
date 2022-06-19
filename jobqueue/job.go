package jobqueue

import (
	"fmt"
	"time"
)

// JobStatus is the status of a job
type JobStatus int

const (
	JobStatusPending JobStatus = iota
	JobStatusDispatched
	JobStatusRunning
	JobStatusCancled
	JobStatusFinished
	JobStatusFailed
)

func (s JobStatus) String() string {
	switch s {
	case JobStatusPending:
		return "pending"
	case JobStatusDispatched:
		return "dispatched"
	case JobStatusRunning:
		return "running"
	case JobStatusCancled:
		return "cancled"
	case JobStatusFinished:
		return "finished"
	case JobStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// Job is a struct that holds the information of a job
// AssigneeID:
type Job struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Data         string    `json:"content"`
	Status       JobStatus `json:"status"`
	ProgressData string    `json:"progress_data"`
	Type         string    `json:"type"`
	Owner        string    `json:"owner_id"`
	AssignTo     string    `json:"assign_to"` // if assign_to is empty, it means the job will be assigned randomely
	ResultCode   int       `json:"result_code"`
	ResultData   string    `json:"result_data"`
	ErrorMessage string    `json:"error_message"`
	CreatedAt    time.Time `json:"created_at"`
	ScheduleAt   time.Time `json:"schedule_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Job) CreateTableSQL(tableName string) string {
	return fmt.Sprintf(
		`CREATE TABLE IF NOT EXISTS %s (
			id BIGINT PRIMARY KEY AUTO_RANDOM,
			name VARCHAR(255) NOT NULL,
			data TEXT DEFAULT NULL,
			status INT NOT NULL,
			type VARCHAR(255) NOT NULL DEFAULT '',
			owner VARCHAR(255) NOT NULL DEFAULT '',
			assign_to VARCHAR(255) NOT NULL DEFAULT '',
			progress_data TEXT DEFAULT NULL,
			result_code INT DEFAULT NULL,
			result_data TEXT DEFAULT NULL,
			error_message TEXT DEFAULT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			schedule_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			UNIQUE (name),
			KEY(owner),
			KEY(schedule_at),
			KEY(status),
			KEY(type),
			KEY(assign_to)
		)`, tableName)
}

func NewJob(name, tp, content string) *Job {
	return &Job{
		Name:      name,
		Data:      content,
		Status:    JobStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func NewJobWithID(id int64, tp, name, content string) *Job {
	return &Job{
		ID:        id,
		Name:      name,
		Data:      content,
		Status:    JobStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
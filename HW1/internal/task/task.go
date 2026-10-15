package task

import "time"

type ID int64
type Status string

const (
	Planned    Status = "planned"
	InProgress Status = "in_progress"
	Canceled   Status = "canceled"
	Done       Status = "done"
)

func (s Status) Valid() bool {
	switch s {
	case Planned, InProgress, Canceled, Done:
		return true
	default:
		return false
	}
}

func (s Status) Terminal() bool { return s == Done || s == Canceled }

type Task struct {
	ID       ID        `json:"id"`
	Title    string    `json:"title"`
	Status   Status    `json:"status"`
	Deadline time.Time `json:"deadline"`
}

type Snapshot struct {
	NextID ID     `json:"next_id"`
	Tasks  []Task `json:"tasks"`
}

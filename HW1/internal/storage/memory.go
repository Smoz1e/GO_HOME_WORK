package storage

import "gitlab.ozon.dev/go/classrom/HW1/internal/task"

type Memory struct{ state task.Snapshot }

func NewMemory() *Memory { return &Memory{state: task.Snapshot{NextID: 1}} }

func (m *Memory) Load() (task.Snapshot, error) {
	state := m.state
	state.Tasks = append([]task.Task(nil), state.Tasks...)
	return state, nil
}

func (m *Memory) Save(state task.Snapshot) error {
	state.Tasks = append([]task.Task(nil), state.Tasks...)
	m.state = state
	return nil
}

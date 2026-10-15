package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gitlab.ozon.dev/go/classrom/HW1/internal/task"
)

type File struct{ path string }

func NewFile(path string) *File { return &File{path: path} }

func (f *File) Load() (task.Snapshot, error) {
	data, err := os.ReadFile(f.path)
	if errors.Is(err, os.ErrNotExist) {
		return task.Snapshot{NextID: 1}, nil
	}
	if err != nil {
		return task.Snapshot{}, fmt.Errorf("прочитать файл задач: %w", err)
	}
	var state task.Snapshot
	if err := json.Unmarshal(data, &state); err != nil {
		return task.Snapshot{}, fmt.Errorf("разобрать файл задач: %w", err)
	}
	return state, nil
}

func (f *File) Save(state task.Snapshot) (resultErr error) {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("сериализовать задачи: %w", err)
	}
	dir := filepath.Dir(f.path)
	tmp, err := os.CreateTemp(dir, ".tasks-*.tmp")
	if err != nil {
		return fmt.Errorf("создать временный файл задач: %w", err)
	}
	defer func() {
		if err := os.Remove(tmp.Name()); err != nil && !errors.Is(err, os.ErrNotExist) && resultErr == nil {
			resultErr = fmt.Errorf("удалить временный файл задач: %w", err)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("записать временный файл задач: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("синхронизировать временный файл задач: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("закрыть временный файл задач: %w", err)
	}
	if err := os.Rename(tmp.Name(), f.path); err != nil {
		return fmt.Errorf("заменить файл задач: %w", err)
	}
	return nil
}

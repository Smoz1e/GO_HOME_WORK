package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"gitlab.ozon.dev/go/classrom/HW1/internal/task"
)

type ImportError struct {
	Index int
	Err   error
}

func (e ImportError) Error() string { return fmt.Sprintf("элемент %d: %v", e.Index, e.Err) }
func (e ImportError) Unwrap() error { return e.Err }

type ImportResult struct {
	Added  int
	Errors []ImportError
}

// ImportJSON принимает массив объектов JSON с заголовком, дедлайном
// и необязательным статусом. Каждая корректная задача сохраняется отдельно.
func (s *Service) ImportJSON(data []byte) (ImportResult, error) {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '[' {
		return ImportResult{}, fmt.Errorf("импортировать задачи: %w", errors.New("ожидается массив JSON"))
	}
	var raw []json.RawMessage
	if err := json.Unmarshal(trimmed, &raw); err != nil {
		return ImportResult{}, fmt.Errorf("разобрать JSON для импорта: %w", err)
	}
	result := ImportResult{}
	for i, entry := range raw {
		var input struct {
			Title    string      `json:"title"`
			Deadline string      `json:"deadline"`
			Status   task.Status `json:"status"`
		}
		if err := json.Unmarshal(entry, &input); err != nil {
			result.Errors = append(result.Errors, ImportError{Index: i + 1, Err: fmt.Errorf("разобрать задачу: %w", err)})
			continue
		}
		if bytes.Equal(bytes.TrimSpace(entry), []byte("null")) {
			result.Errors = append(result.Errors, ImportError{Index: i + 1, Err: errors.New("задача должна быть объектом")})
			continue
		}
		deadline, err := time.ParseInLocation("2006-01-02", input.Deadline, s.now().Location())
		if err != nil {
			result.Errors = append(result.Errors, ImportError{Index: i + 1, Err: fmt.Errorf("неверный дедлайн: %w", err)})
			continue
		}
		if err := s.validate(input.Title, deadline); err != nil {
			result.Errors = append(result.Errors, ImportError{Index: i + 1, Err: fmt.Errorf("некорректная задача: %w", err)})
			continue
		}
		if input.Status == "" {
			input.Status = task.Planned
		}
		if !input.Status.Valid() {
			result.Errors = append(result.Errors, ImportError{Index: i + 1, Err: errors.New("некорректный статус")})
			continue
		}
		if s.state.NextID == math.MaxInt64 {
			return result, fmt.Errorf("импортировать элемент %d: %w", i+1, ErrIDExhausted)
		}
		item := task.Task{ID: s.state.NextID, Title: strings.TrimSpace(input.Title), Deadline: deadline, Status: input.Status}
		state := s.state
		state.NextID++
		state.Tasks = append(append([]task.Task(nil), state.Tasks...), item)
		if err := s.save(state); err != nil {
			return result, fmt.Errorf("импортировать элемент %d: %w", i+1, err)
		}
		result.Added++
	}
	return result, nil
}

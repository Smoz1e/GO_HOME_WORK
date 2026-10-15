package service

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"gitlab.ozon.dev/go/classrom/HW1/internal/task"
)

var ErrNotFound = errors.New("задача не найдена")
var ErrIDExhausted = errors.New("закончились доступные идентификаторы")

type Store interface {
	Load() (task.Snapshot, error)
	Save(task.Snapshot) error
}

type Service struct {
	store Store
	state task.Snapshot
	now   func() time.Time
}

func New(store Store, now func() time.Time) (*Service, error) {
	if store == nil || now == nil {
		return nil, fmt.Errorf("создать сервис: %w", errors.New("хранилище и источник времени обязательны"))
	}
	state, err := store.Load()
	if err != nil {
		return nil, fmt.Errorf("загрузить задачи: %w", err)
	}
	if state.NextID < 1 {
		return nil, fmt.Errorf("загрузить задачи: %w", errors.New("некорректный следующий идентификатор"))
	}
	seen := make(map[task.ID]bool, len(state.Tasks))
	for _, item := range state.Tasks {
		if item.ID < 1 || item.ID >= state.NextID || seen[item.ID] || !item.Status.Valid() || item.Deadline.IsZero() || strings.TrimSpace(item.Title) == "" {
			return nil, fmt.Errorf("загрузить задачи: %w", errors.New("некорректные данные задач в файле"))
		}
		seen[item.ID] = true
	}
	return &Service{store: store, state: state, now: now}, nil
}

func dateKey(value time.Time) string { return value.Format("2006-01-02") }

func (s *Service) validate(title string, deadline time.Time) error {
	if strings.TrimSpace(title) == "" {
		return errors.New("заголовок не может быть пустым")
	}
	if deadline.IsZero() {
		return errors.New("дедлайн обязателен")
	}
	if dateKey(deadline) < dateKey(s.now()) {
		return errors.New("дедлайн не может быть в прошлом")
	}
	return nil
}

func (s *Service) save(state task.Snapshot) error {
	if err := s.store.Save(state); err != nil {
		return fmt.Errorf("сохранить задачи: %w", err)
	}
	s.state = state
	return nil
}

func (s *Service) Create(title string, deadline time.Time) (task.Task, error) {
	if err := s.validate(title, deadline); err != nil {
		return task.Task{}, fmt.Errorf("создать задачу: %w", err)
	}
	if s.state.NextID == math.MaxInt64 {
		return task.Task{}, fmt.Errorf("создать задачу: %w", ErrIDExhausted)
	}
	item := task.Task{ID: s.state.NextID, Title: strings.TrimSpace(title), Status: task.Planned, Deadline: deadline}
	state := s.state
	state.NextID++
	state.Tasks = append(append([]task.Task(nil), state.Tasks...), item)
	if err := s.save(state); err != nil {
		return task.Task{}, fmt.Errorf("создать задачу: %w", err)
	}
	return item, nil
}

func (s *Service) index(id task.ID) (int, error) {
	for i, item := range s.state.Tasks {
		if item.ID == id {
			return i, nil
		}
	}
	return -1, fmt.Errorf("задача %d: %w", id, ErrNotFound)
}

func (s *Service) Get(id task.ID) (task.Task, error) {
	i, err := s.index(id)
	if err != nil {
		return task.Task{}, fmt.Errorf("получить задачу: %w", err)
	}
	return s.state.Tasks[i], nil
}

func (s *Service) UpdateStatus(id task.ID, status task.Status) (task.Task, error) {
	i, err := s.index(id)
	if err != nil {
		return task.Task{}, fmt.Errorf("изменить статус: %w", err)
	}
	item := s.state.Tasks[i]
	if !status.Valid() {
		return task.Task{}, fmt.Errorf("изменить статус: %w", errors.New("некорректный статус"))
	}
	if item.Status.Terminal() {
		return task.Task{}, fmt.Errorf("изменить статус: %w", errors.New("нельзя изменить завершённую или отменённую задачу"))
	}
	if item.Status == status {
		return task.Task{}, fmt.Errorf("изменить статус: %w", errors.New("у задачи уже установлен этот статус"))
	}
	item.Status = status
	state := s.state
	state.Tasks = append([]task.Task(nil), state.Tasks...)
	state.Tasks[i] = item
	if err := s.save(state); err != nil {
		return task.Task{}, fmt.Errorf("изменить статус: %w", err)
	}
	return item, nil
}

func (s *Service) Edit(id task.ID, title *string, deadline *time.Time) (task.Task, error) {
	i, err := s.index(id)
	if err != nil {
		return task.Task{}, fmt.Errorf("редактировать задачу: %w", err)
	}
	item := s.state.Tasks[i]
	if item.Status.Terminal() {
		return task.Task{}, fmt.Errorf("редактировать задачу: %w", errors.New("нельзя редактировать завершённую или отменённую задачу"))
	}
	if title == nil && deadline == nil {
		return task.Task{}, fmt.Errorf("редактировать задачу: %w", errors.New("укажите заголовок или дедлайн"))
	}
	if title != nil {
		item.Title = strings.TrimSpace(*title)
	}
	if deadline != nil {
		item.Deadline = *deadline
	}
	if err := s.validate(item.Title, item.Deadline); err != nil {
		return task.Task{}, fmt.Errorf("редактировать задачу: %w", err)
	}
	state := s.state
	state.Tasks = append([]task.Task(nil), state.Tasks...)
	state.Tasks[i] = item
	if err := s.save(state); err != nil {
		return task.Task{}, fmt.Errorf("редактировать задачу: %w", err)
	}
	return item, nil
}

func (s *Service) Delete(id task.ID) error {
	i, err := s.index(id)
	if err != nil {
		return fmt.Errorf("удалить задачу: %w", err)
	}
	state := s.state
	state.Tasks = append([]task.Task(nil), state.Tasks[:i]...)
	state.Tasks = append(state.Tasks, s.state.Tasks[i+1:]...)
	if err := s.save(state); err != nil {
		return fmt.Errorf("удалить задачу: %w", err)
	}
	return nil
}

type ListOptions struct {
	Status  task.Status
	Overdue bool
	Search  string
	Page    int
	Limit   int
}

type ListResult struct {
	Tasks []task.Task
	Total int
	Pages int
	Page  int
	Limit int
}

func (s *Service) List(options ListOptions) (ListResult, error) {
	if options.Page == 0 {
		options.Page = 1
	}
	if options.Limit == 0 {
		options.Limit = 20
	}
	if options.Page < 1 || options.Limit < 1 {
		return ListResult{}, fmt.Errorf("получить список задач: %w", errors.New("номер страницы и размер страницы должны быть положительными"))
	}
	if options.Status != "" && !options.Status.Valid() {
		return ListResult{}, fmt.Errorf("получить список задач: %w", errors.New("неверный фильтр статуса"))
	}
	var filtered []task.Task
	needle := strings.ToLower(options.Search)
	today := dateKey(s.now())
	for _, item := range s.state.Tasks {
		if options.Status != "" && item.Status != options.Status {
			continue
		}
		if options.Overdue && (item.Status.Terminal() || dateKey(item.Deadline) >= today) {
			continue
		}
		if !strings.Contains(strings.ToLower(item.Title), needle) {
			continue
		}
		filtered = append(filtered, item)
	}
	sort.Slice(filtered, func(i, j int) bool {
		left, right := dateKey(filtered[i].Deadline), dateKey(filtered[j].Deadline)
		if left == right {
			return filtered[i].ID < filtered[j].ID
		}
		return left < right
	})
	total := len(filtered)
	pages := total / options.Limit
	if total%options.Limit != 0 {
		pages++
	}
	if options.Page > pages || total == 0 {
		return ListResult{Total: total, Pages: pages, Page: options.Page, Limit: options.Limit}, nil
	}
	start := (options.Page - 1) * options.Limit
	end := start + options.Limit
	if end > total {
		end = total
	}
	return ListResult{Tasks: filtered[start:end], Total: total, Pages: pages, Page: options.Page, Limit: options.Limit}, nil
}

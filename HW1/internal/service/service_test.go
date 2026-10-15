package service_test

import (
	"encoding/json"
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"gitlab.ozon.dev/go/classrom/HW1/internal/service"
	"gitlab.ozon.dev/go/classrom/HW1/internal/storage"
	"gitlab.ozon.dev/go/classrom/HW1/internal/task"
)

func testService(t *testing.T) *service.Service {
	t.Helper()
	s, err := service.New(storage.NewMemory(), func() time.Time {
		return time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func date(day int) time.Time { return time.Date(2026, 10, day, 0, 0, 0, 0, time.UTC) }

func TestCreateValidationAndIDsAfterDelete(t *testing.T) {
	s := testService(t)
	if _, err := s.Create("  ", date(15)); err == nil {
		t.Fatal("empty title accepted")
	}
	if _, err := s.Create("past", date(13)); err == nil {
		t.Fatal("past deadline accepted")
	}
	first, err := s.Create("first", date(14))
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != 1 || first.Status != task.Planned {
		t.Fatalf("unexpected first task: %+v", first)
	}
	if err := s.Delete(first.ID); err != nil {
		t.Fatal(err)
	}
	second, err := s.Create("second", date(15))
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != 2 {
		t.Fatalf("reused ID: %d", second.ID)
	}
	if _, err := s.Get(first.ID); !errors.Is(err, service.ErrNotFound) {
		t.Fatalf("missing task: %v", err)
	}
}

func TestStatusAndEditRules(t *testing.T) {
	s := testService(t)
	created, err := s.Create("draft", date(15))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateStatus(created.ID, task.Planned); err == nil {
		t.Fatal("accepted same status")
	}
	if _, err := s.UpdateStatus(created.ID, task.Status("invalid")); err == nil {
		t.Fatal("accepted invalid status")
	}
	title := "updated"
	if _, err := s.Edit(created.ID, &title, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateStatus(created.ID, task.Done); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateStatus(created.ID, task.InProgress); err == nil {
		t.Fatal("changed done task")
	}
	if _, err := s.Edit(created.ID, &title, nil); err == nil {
		t.Fatal("edited done task")
	}
	if err := s.Delete(created.ID); err != nil {
		t.Fatal(err)
	}
}

func TestListFiltersSortsAndPaginates(t *testing.T) {
	s := testService(t)
	for _, item := range []struct {
		title string
		day   int
	}{{"Alpha", 16}, {"beta", 15}, {"ALPHABET", 15}} {
		if _, err := s.Create(item.title, date(item.day)); err != nil {
			t.Fatal(err)
		}
	}
	result, err := s.List(service.ListOptions{Page: 1, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 3 || result.Pages != 2 || len(result.Tasks) != 2 || result.Tasks[0].ID != 2 || result.Tasks[1].ID != 3 {
		t.Fatalf("unexpected first page: %+v", result)
	}
	result, err = s.List(service.ListOptions{Search: "alpha", Page: 2, Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 2 || result.Pages != 2 || len(result.Tasks) != 1 || result.Tasks[0].ID != 1 {
		t.Fatalf("unexpected search page: %+v", result)
	}
	if _, err := s.List(service.ListOptions{Page: -1}); err == nil {
		t.Fatal("accepted negative page")
	}
}

func TestImportContinuesAfterBadItems(t *testing.T) {
	s := testService(t)
	input := []byte(`[{"title":"one","deadline":"2026-10-15"},{"title":"","deadline":"2026-10-16"},{"title":"three","deadline":"2026-10-17","status":"done"}]`)
	result, err := s.ImportJSON(input)
	if err != nil {
		t.Fatal(err)
	}
	if result.Added != 2 || len(result.Errors) != 1 || result.Errors[0].Index != 2 {
		t.Fatalf("unexpected import: %+v", result)
	}
	list, err := s.List(service.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if list.Total != 2 || list.Tasks[1].Status != task.Done {
		t.Fatalf("unexpected tasks: %+v", list)
	}
	if _, err := s.ImportJSON([]byte(`{}`)); err == nil {
		t.Fatal("accepted non-array JSON")
	}
	if _, err := json.Marshal(result); err != nil {
		t.Fatal(err)
	}
}

func TestOverdueExcludesFinishedTasks(t *testing.T) {
	s := testService(t)
	input := []byte(`[{"title":"late","deadline":"2026-10-13"},{"title":"finished","deadline":"2026-10-13","status":"done"}]`)

	store := storage.NewMemory()
	earlier, err := service.New(store, func() time.Time { return date(12) })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := earlier.ImportJSON(input); err != nil {
		t.Fatal(err)
	}
	s, err = service.New(store, func() time.Time { return date(14) })
	if err != nil {
		t.Fatal(err)
	}
	result, err := s.List(service.ListOptions{Overdue: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 1 || !strings.EqualFold(result.Tasks[0].Title, "late") {
		t.Fatalf("unexpected overdue list: %+v", result)
	}
}

func TestCalendarDatesStaySortedAcrossTimeZonesAndLargePages(t *testing.T) {
	store := storage.NewMemory()
	s, err := service.New(store, func() time.Time { return date(14) })
	if err != nil {
		t.Fatal(err)
	}
	west := time.Date(2026, 10, 16, 0, 0, 0, 0, time.FixedZone("west", -8*3600))
	east := time.Date(2026, 10, 16, 0, 0, 0, 0, time.FixedZone("east", 8*3600))
	first, err := s.Create("first", west)
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.Create("second", east)
	if err != nil {
		t.Fatal(err)
	}
	result, err := s.List(service.ListOptions{Page: 1, Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if result.Tasks[0].ID != first.ID || result.Tasks[1].ID != second.ID {
		t.Fatalf("equal calendar dates sorted by instant: %+v", result.Tasks)
	}
	result, err = s.List(service.ListOptions{Page: int(^uint(0) >> 1), Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Tasks) != 0 || result.Total != 2 {
		t.Fatalf("unexpected out-of-range page: %+v", result)
	}
}

func TestExhaustedIDSpaceDoesNotWriteInvalidState(t *testing.T) {
	store := storage.NewMemory()
	if err := store.Save(task.Snapshot{NextID: math.MaxInt64}); err != nil {
		t.Fatal(err)
	}
	s, err := service.New(store, func() time.Time { return date(14) })
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Create("new", date(15)); err == nil {
		t.Fatal("accepted exhausted ID space")
	}
	result, err := s.ImportJSON([]byte(`[{"title":"new","deadline":"2026-10-15"}]`))
	if err == nil || result.Added != 0 {
		t.Fatalf("import accepted exhausted ID space: %+v, %v", result, err)
	}
	state, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if state.NextID != math.MaxInt64 || len(state.Tasks) != 0 {
		t.Fatalf("state changed: %+v", state)
	}
}

type failOnceStore struct {
	state  task.Snapshot
	failed bool
}

func (s *failOnceStore) Load() (task.Snapshot, error) { return s.state, nil }
func (s *failOnceStore) Save(state task.Snapshot) error {
	if !s.failed {
		s.failed = true
		return errors.New("disk unavailable")
	}
	s.state = state
	return nil
}

func TestImportStopsOnStorageFailure(t *testing.T) {
	store := &failOnceStore{state: task.Snapshot{NextID: 1}}
	s, err := service.New(store, func() time.Time { return date(14) })
	if err != nil {
		t.Fatal(err)
	}
	result, err := s.ImportJSON([]byte(`[{"title":"first","deadline":"2026-10-15"},{"title":"second","deadline":"2026-10-16"}]`))
	if err == nil || result.Added != 0 {
		t.Fatalf("import continued after storage failure: %+v, %v", result, err)
	}
	list, err := s.List(service.ListOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if list.Total != 0 {
		t.Fatalf("imported another item after failure: %+v", list)
	}
}

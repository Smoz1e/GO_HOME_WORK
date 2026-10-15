package storage_test

import (
	"path/filepath"
	"testing"
	"time"

	"gitlab.ozon.dev/go/classrom/HW1/internal/service"
	"gitlab.ozon.dev/go/classrom/HW1/internal/storage"
)

func TestFilePersistsTasksAndNextID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.json")
	now := func() time.Time { return time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC) }
	first, err := service.New(storage.NewFile(path), now)
	if err != nil {
		t.Fatal(err)
	}
	created, err := first.Create("persisted", time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Delete(created.ID); err != nil {
		t.Fatal(err)
	}
	second, err := service.New(storage.NewFile(path), now)
	if err != nil {
		t.Fatal(err)
	}
	other, err := second.Create("next", time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if other.ID != 2 {
		t.Fatalf("ID after restart = %d", other.ID)
	}
	third, err := service.New(storage.NewFile(path), now)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := third.Get(other.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Title != "next" {
		t.Fatalf("loaded wrong task: %+v", loaded)
	}
}

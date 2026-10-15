package cli_test

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"gitlab.ozon.dev/go/classrom/HW1/internal/cli"
	"gitlab.ozon.dev/go/classrom/HW1/internal/service"
	"gitlab.ozon.dev/go/classrom/HW1/internal/storage"
)

func TestInteractiveCommandsAndRecovery(t *testing.T) {
	s, err := service.New(storage.NewMemory(), func() time.Time { return time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC) })
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	in := strings.NewReader("add \"Buy milk\" 2026-10-15\nget 1\nstatus 1 done\nedit 1 --title nope\nlist --page 1 --limit 1\nexit\n")
	if err := cli.Run(in, &out, s); err != nil {
		t.Fatal(err)
	}
	for _, piece := range []string{"Buy milk", "done", "нельзя редактировать", "Всего задач: 1", "Всего страниц: 1"} {
		if !strings.Contains(out.String(), piece) {
			t.Fatalf("missing %q in %q", piece, out.String())
		}
	}
}

func TestExplicitZeroPaginationIsRejected(t *testing.T) {
	s, err := service.New(storage.NewMemory(), time.Now)
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := cli.Run(strings.NewReader("list --page 0\nlist --limit 0\nexit\n"), &out, s); err != nil {
		t.Fatal(err)
	}
	if strings.Count(out.String(), "номер страницы и размер страницы должны быть положительными") != 2 {
		t.Fatalf("zero pagination accepted: %q", out.String())
	}
}

func TestHelpAndErrorsAreInRussian(t *testing.T) {
	s, err := service.New(storage.NewMemory(), func() time.Time { return time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC) })
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	in := strings.NewReader("help\nget 999\nadd \"\" 2026-10-15\nadd milk 2026-10-13\nlist --status wrong\nunknown\nexit\n")
	if err := cli.Run(in, &out, s); err != nil {
		t.Fatal(err)
	}
	for _, piece := range []string{
		"Доступные команды:",
		"Ошибка: получить задачу",
		"задача не найдена",
		"заголовок не может быть пустым",
		"дедлайн не может быть в прошлом",
		"неверный фильтр статуса",
		"неизвестная команда",
	} {
		if !strings.Contains(out.String(), piece) {
			t.Fatalf("missing %q in %q", piece, out.String())
		}
	}
}

func TestLibraryErrorsAreShownInRussian(t *testing.T) {
	s, err := service.New(storage.NewMemory(), func() time.Time { return time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC) })
	if err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	in := strings.NewReader("get abc\nadd milk wrong-date\nlist --page abc\nimport /no-such-task-file.json\nexit\n")
	if err := cli.Run(in, &out, s); err != nil {
		t.Fatal(err)
	}
	for _, piece := range []string{"ожидается целое число", "ожидается дата", "файл", "не найден"} {
		if !strings.Contains(out.String(), piece) {
			t.Fatalf("missing %q in %q", piece, out.String())
		}
	}
	for _, piece := range []string{"strconv", "parsing time", "no such file", "open /"} {
		if strings.Contains(out.String(), piece) {
			t.Fatalf("English detail %q in %q", piece, out.String())
		}
	}
}

func TestFileReplacementErrorIsShownInRussian(t *testing.T) {
	err := fmt.Errorf("сохранить задачи: %w", &os.LinkError{Op: "rename", Old: "old.json", New: "tasks.json", Err: os.ErrPermission})
	message := cli.Message(err)
	if !strings.Contains(message, "нет доступа") || strings.Contains(message, "rename") || strings.Contains(message, "permission denied") {
		t.Fatalf("unexpected error message: %q", message)
	}
}

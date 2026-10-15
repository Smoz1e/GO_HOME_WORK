package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"gitlab.ozon.dev/go/classrom/HW1/internal/service"
	"gitlab.ozon.dev/go/classrom/HW1/internal/task"
)

const help = `Доступные команды:
  help                                      показать справку
  add "заголовок" ГГГГ-ММ-ДД                 добавить задачу
  list [--status СТАТУС] [--overdue] [--search ТЕКСТ] [--page N] [--limit N]
                                            показать и отфильтровать задачи
  get ID                                    показать задачу
  status ID СТАТУС                          изменить статус
  edit ID [--title "заголовок"] [--deadline ГГГГ-ММ-ДД]
                                            изменить заголовок и/или дедлайн
  delete ID                                 удалить задачу
  import ПУТЬ                               импортировать массив задач из JSON
  exit                                      выйти
Статусы: planned, in_progress, canceled, done
Объект для импорта: {"title":"Пример","deadline":"2030-10-20","status":"planned"}`

func Run(in io.Reader, out io.Writer, s *service.Service) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 4096), 1024*1024)
	for scanner.Scan() {
		args, err := split(scanner.Text())
		if err != nil {
			if _, writeErr := fmt.Fprintf(out, "Ошибка: %s\n", Message(err)); writeErr != nil {
				return fmt.Errorf("вывести ошибку: %w", writeErr)
			}
			continue
		}
		if len(args) == 0 {
			continue
		}
		if args[0] == "exit" {
			if len(args) != 1 {
				err = fmt.Errorf("команда exit не принимает аргументы")
			} else {
				return nil
			}
		} else {
			err = execute(args, out, s)
		}
		if err != nil {
			if _, writeErr := fmt.Fprintf(out, "Ошибка: %s\n", Message(err)); writeErr != nil {
				return fmt.Errorf("вывести ошибку: %w", writeErr)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("прочитать команды: %w", err)
	}
	return nil
}

func printTask(out io.Writer, item task.Task) error {
	if _, err := fmt.Fprintf(out, "%d | %s | %s | %s\n", item.ID, item.Title, item.Status, item.Deadline.Format("2006-01-02")); err != nil {
		return fmt.Errorf("вывести задачу: %w", err)
	}
	return nil
}

func parseID(raw string) (task.ID, error) {
	n, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("некорректный ID: %w", err)
	}
	if n < 1 {
		return 0, fmt.Errorf("некорректный ID: должен быть положительным")
	}
	return task.ID(n), nil
}

func parseDate(raw string) (time.Time, error) {
	value, err := time.ParseInLocation("2006-01-02", raw, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("некорректная дата (формат ГГГГ-ММ-ДД): %w", err)
	}
	return value, nil
}

func valueFlags(args []string, allowed map[string]bool) (map[string]string, error) {
	values := make(map[string]string)
	for i := 0; i < len(args); i++ {
		flag := args[i]
		if !allowed[flag] {
			return nil, fmt.Errorf("неизвестный параметр %q", flag)
		}
		if _, exists := values[flag]; exists {
			return nil, fmt.Errorf("параметр %q указан повторно", flag)
		}
		if flag == "--overdue" {
			values[flag] = "true"
			continue
		}
		if i+1 >= len(args) {
			return nil, fmt.Errorf("не указано значение параметра %s", flag)
		}
		i++
		values[flag] = args[i]
	}
	return values, nil
}

func execute(args []string, out io.Writer, s *service.Service) error {
	switch args[0] {
	case "help":
		if len(args) != 1 {
			return fmt.Errorf("команда help не принимает аргументы")
		}
		if _, err := fmt.Fprintln(out, help); err != nil {
			return fmt.Errorf("показать справку: %w", err)
		}
		return nil
	case "add":
		if len(args) != 3 {
			return fmt.Errorf("использование: add \"заголовок\" ГГГГ-ММ-ДД")
		}
		deadline, err := parseDate(args[2])
		if err != nil {
			return fmt.Errorf("добавить задачу: %w", err)
		}
		item, err := s.Create(args[1], deadline)
		if err != nil {
			return fmt.Errorf("добавить задачу: %w", err)
		}
		if err := printTask(out, item); err != nil {
			return fmt.Errorf("добавить задачу: %w", err)
		}
		return nil
	case "get":
		if len(args) != 2 {
			return fmt.Errorf("использование: get ID")
		}
		id, err := parseID(args[1])
		if err != nil {
			return fmt.Errorf("получить задачу: %w", err)
		}
		item, err := s.Get(id)
		if err != nil {
			return fmt.Errorf("получить задачу: %w", err)
		}
		if err := printTask(out, item); err != nil {
			return fmt.Errorf("получить задачу: %w", err)
		}
		return nil
	case "delete":
		if len(args) != 2 {
			return fmt.Errorf("использование: delete ID")
		}
		id, err := parseID(args[1])
		if err != nil {
			return fmt.Errorf("удалить задачу: %w", err)
		}
		if err := s.Delete(id); err != nil {
			return fmt.Errorf("удалить задачу: %w", err)
		}
		if _, err := fmt.Fprintf(out, "Задача %d удалена\n", id); err != nil {
			return fmt.Errorf("удалить задачу: %w", err)
		}
		return nil
	case "status":
		if len(args) != 3 {
			return fmt.Errorf("использование: status ID СТАТУС")
		}
		id, err := parseID(args[1])
		if err != nil {
			return fmt.Errorf("изменить статус: %w", err)
		}
		item, err := s.UpdateStatus(id, task.Status(args[2]))
		if err != nil {
			return fmt.Errorf("изменить статус: %w", err)
		}
		if err := printTask(out, item); err != nil {
			return fmt.Errorf("изменить статус: %w", err)
		}
		return nil
	case "edit":
		if len(args) < 4 {
			return fmt.Errorf("использование: edit ID [--title \"заголовок\"] [--deadline ГГГГ-ММ-ДД]")
		}
		id, err := parseID(args[1])
		if err != nil {
			return fmt.Errorf("редактировать задачу: %w", err)
		}
		values, err := valueFlags(args[2:], map[string]bool{"--title": true, "--deadline": true})
		if err != nil {
			return fmt.Errorf("редактировать задачу: %w", err)
		}
		var title *string
		var deadline *time.Time
		if raw, ok := values["--title"]; ok {
			title = &raw
		}
		if raw, ok := values["--deadline"]; ok {
			parsed, parseErr := parseDate(raw)
			if parseErr != nil {
				return fmt.Errorf("редактировать задачу: %w", parseErr)
			}
			deadline = &parsed
		}
		item, err := s.Edit(id, title, deadline)
		if err != nil {
			return fmt.Errorf("редактировать задачу: %w", err)
		}
		if err := printTask(out, item); err != nil {
			return fmt.Errorf("редактировать задачу: %w", err)
		}
		return nil
	case "list":
		values, err := valueFlags(args[1:], map[string]bool{"--status": true, "--overdue": true, "--search": true, "--page": true, "--limit": true})
		if err != nil {
			return fmt.Errorf("получить список задач: %w", err)
		}
		options := service.ListOptions{Status: task.Status(values["--status"]), Overdue: values["--overdue"] == "true", Search: values["--search"]}
		if raw, ok := values["--page"]; ok {
			options.Page, err = strconv.Atoi(raw)
			if err != nil {
				return fmt.Errorf("некорректный номер страницы: %w", err)
			}
			if options.Page < 1 {
				return fmt.Errorf("получить список задач: номер страницы и размер страницы должны быть положительными")
			}
		}
		if raw, ok := values["--limit"]; ok {
			options.Limit, err = strconv.Atoi(raw)
			if err != nil {
				return fmt.Errorf("некорректный размер страницы: %w", err)
			}
			if options.Limit < 1 {
				return fmt.Errorf("получить список задач: номер страницы и размер страницы должны быть положительными")
			}
		}
		result, err := s.List(options)
		if err != nil {
			return fmt.Errorf("получить список задач: %w", err)
		}
		for _, item := range result.Tasks {
			if err := printTask(out, item); err != nil {
				return fmt.Errorf("получить список задач: %w", err)
			}
		}
		if _, err := fmt.Fprintf(out, "Всего задач: %d | Всего страниц: %d | Страница: %d\n", result.Total, result.Pages, result.Page); err != nil {
			return fmt.Errorf("получить список задач: %w", err)
		}
		return nil
	case "import":
		if len(args) != 2 {
			return fmt.Errorf("использование: import ПУТЬ")
		}
		data, err := os.ReadFile(args[1])
		if err != nil {
			return fmt.Errorf("прочитать файл для импорта: %w", err)
		}
		result, err := s.ImportJSON(data)
		if err != nil {
			return fmt.Errorf("импортировать задачи: %w", err)
		}
		if _, err := fmt.Fprintf(out, "Импортировано задач: %d\n", result.Added); err != nil {
			return fmt.Errorf("показать результат импорта: %w", err)
		}
		for _, itemErr := range result.Errors {
			if _, err := fmt.Fprintf(out, "Ошибка: %s\n", Message(itemErr)); err != nil {
				return fmt.Errorf("показать ошибку импорта: %w", err)
			}
		}
		return nil
	default:
		return fmt.Errorf("неизвестная команда %q; используйте help", strings.TrimSpace(args[0]))
	}
}

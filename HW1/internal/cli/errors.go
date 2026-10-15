package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Message сохраняет контекст ошибки, но показывает ошибки стандартной
// библиотеки понятными сообщениями на русском языке.
func Message(err error) string {
	var parts []string
	for err != nil {
		if text, known := libraryError(err); known {
			parts = append(parts, text)
			break
		}
		cause := errors.Unwrap(err)
		if cause == nil {
			parts = append(parts, err.Error())
			break
		}
		suffix := ": " + cause.Error()
		if !strings.HasSuffix(err.Error(), suffix) {
			parts = append(parts, err.Error())
			break
		}
		context := strings.TrimSuffix(err.Error(), suffix)
		if context != "" && (len(parts) == 0 || parts[len(parts)-1] != context) {
			parts = append(parts, context)
		}
		err = cause
	}
	return strings.Join(parts, ": ")
}

func libraryError(err error) (string, bool) {
	switch value := err.(type) {
	case *strconv.NumError:
		return "ожидается целое число", true
	case *time.ParseError:
		return "ожидается дата в формате ГГГГ-ММ-ДД", true
	case *json.SyntaxError:
		return fmt.Sprintf("некорректный синтаксис JSON в позиции %d", value.Offset), true
	case *json.UnmarshalTypeError:
		return "неверный тип значения в JSON", true
	case *os.PathError:
		if errors.Is(value, os.ErrNotExist) {
			return fmt.Sprintf("файл %q не найден", value.Path), true
		}
		if errors.Is(value, os.ErrPermission) {
			return fmt.Sprintf("нет доступа к файлу %q", value.Path), true
		}
		return fmt.Sprintf("ошибка доступа к файлу %q", value.Path), true
	case *os.LinkError:
		if errors.Is(value, os.ErrNotExist) {
			return fmt.Sprintf("файл %q не найден", value.New), true
		}
		if errors.Is(value, os.ErrPermission) {
			return fmt.Sprintf("нет доступа к файлу %q", value.New), true
		}
		return fmt.Sprintf("не удалось заменить файл %q", value.New), true
	default:
		return "", false
	}
}

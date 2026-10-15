package cli

import (
	"errors"
	"strings"
	"unicode"
)

func split(line string) ([]string, error) {
	var args []string
	var current strings.Builder
	var quote rune
	escaped := false
	active := false
	for _, r := range line {
		switch {
		case escaped:
			current.WriteRune(r)
			escaped = false
			active = true
		case r == '\\' && quote != '\'':
			escaped = true
			active = true
		case quote != 0 && r == quote:
			quote = 0
		case quote != 0:
			current.WriteRune(r)
		case r == '"' || r == '\'':
			quote = r
			active = true
		case unicode.IsSpace(r):
			if active {
				args = append(args, current.String())
				current.Reset()
				active = false
			}
		default:
			current.WriteRune(r)
			active = true
		}
	}
	if escaped {
		return nil, errors.New("незавершённое экранирование")
	}
	if quote != 0 {
		return nil, errors.New("незакрытая кавычка")
	}
	if active {
		args = append(args, current.String())
	}
	return args, nil
}

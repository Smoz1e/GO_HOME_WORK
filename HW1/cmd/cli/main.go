package main

import (
	"fmt"
	"os"
	"time"

	"gitlab.ozon.dev/go/classrom/HW1/internal/cli"
	"gitlab.ozon.dev/go/classrom/HW1/internal/service"
	"gitlab.ozon.dev/go/classrom/HW1/internal/storage"
)

func main() {
	path := os.Getenv("TASKS_FILE")
	if path == "" {
		path = "tasks.json"
	}
	store := storage.NewFile(path)
	s, err := service.New(store, time.Now)
	if err == nil {
		err = cli.Run(os.Stdin, os.Stdout, s)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "Ошибка:", cli.Message(err))
		os.Exit(1)
	}
}

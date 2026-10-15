package main

import (
	"bytes"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("укажите файлы Go")
	}
	for _, path := range os.Args[1:] {
		info, err := os.Stat(path)
		if err != nil {
			log.Fatal(err)
		}
		set := token.NewFileSet()
		file, err := parser.ParseFile(set, path, nil, 0)
		if err != nil {
			log.Fatal(err)
		}
		var output bytes.Buffer
		if err := format.Node(&output, set, file); err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(path, output.Bytes(), info.Mode().Perm()); err != nil {
			log.Fatal(err)
		}
	}
}

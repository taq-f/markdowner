package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	zglob "github.com/mattn/go-zglob"
	"github.com/taq-f/markdowner/renderer"
	"github.com/taq-f/markdowner/style"
	"github.com/taq-f/markdowner/template"
)

func main() {
	argInputFile := flag.String("f", "", "help text")
	argOutDir := flag.String("o", "", "help text")
	flag.Parse()

	files, err := getTargetFiles(*argInputFile)
	if err != nil {
		fmt.Println("error", err)
		return
	}

	template := template.Get()
	style := style.Get()

	for _, f := range files {
		err = renderer.Render(template, style, f, *argOutDir, *argInputFile)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func getTargetFiles(path string) ([]string, error) {
	f, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	mode := f.Mode()
	if !mode.IsDir() {
		return []string{path}, nil
	}

	p := filepath.Join(path, "**", "*.md")
	files, err := zglob.Glob(p)
	if err != nil {
		return nil, err
	}

	return files, nil
}

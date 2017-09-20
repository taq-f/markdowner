package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	zglob "github.com/mattn/go-zglob"
	"github.com/taq-f/markdowner/renderer"
	"github.com/taq-f/markdowner/style"
	"github.com/taq-f/markdowner/template"
)

func main() {
	argInputFile := flag.String("f", "", "")
	argOutDir := flag.String("o", "", "")
	argImageInline := flag.Bool("i", false, "")
	flag.Parse()

	// input file / directory
	// if not specified, current directory
	var input string
	if *argInputFile == "" {
		input, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	} else {
		input = cleanPath(*argInputFile)
	}

	// base directory of input
	var baseDir string
	if *argInputFile == "" {
		baseDir = input
	} else {
		if isDir(input) {
			baseDir = input
		} else {
			baseDir = filepath.Dir(input)
		}
	}

	// output directory
	// if not specified, same as input directory
	var outDir string
	if *argOutDir == "" {
		if isDir(input) {
			outDir = input
		} else {
			outDir = filepath.Dir(input)
		}
	} else {
		outDir = cleanPath(*argOutDir)
	}

	r := renderer.Renderer{
		ImageInline: *argImageInline,
		Template:    template.Get(),
		Style:       style.Get(),
		OutDir:      outDir,
		BaseDir:     baseDir}

	files, err := getTargetFiles(input)
	if err != nil {
		fmt.Println("error", err)
		return
	}

	wait := new(sync.WaitGroup)
	wait.Add(len(files))

	for _, f := range files {
		go func(file string) {
			err = r.Render(file)
			if err != nil {
				fmt.Println(err)
			}
			wait.Done()
		}(f)
	}
	wait.Wait()
}

// collect markdown files from the path specified.
// if the path is a file, return only that file.
// if the path is a directory, return all markdown files under it (recursively).
func getTargetFiles(path string) ([]string, error) {
	if !isDir(path) {
		return []string{path}, nil
	}

	globStr := filepath.Join(path, "**", "*.md")
	files, err := zglob.Glob(globStr)
	if err != nil {
		return nil, err
	}

	return files, nil
}

// see if specified path is a directory or file.
// be careful the path exists and can be read, since this function won't
// return any errors.
func isDir(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		return false
	}

	mode := f.Mode()
	return mode.IsDir()
}

func cleanPath(path string) string {
	return filepath.Join(path)
}

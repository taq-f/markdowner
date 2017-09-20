package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-fsnotify/fsnotify"
	zglob "github.com/mattn/go-zglob"
	"github.com/taq-f/markdowner/renderer"
	"github.com/taq-f/markdowner/style"
	"github.com/taq-f/markdowner/template"
)

func main() {
	argInputFile := flag.String("f", "", "")
	argOutDir := flag.String("o", "", "")
	argImageInline := flag.Bool("i", false, "")
	argWatch := flag.Bool("w", false, "")
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
		log.Fatalln()
	}

	wait := new(sync.WaitGroup)
	wait.Add(len(files))

	for _, f := range files {
		go func(file string) {
			err = r.Render(file)
			if err != nil {
				log.Println(err)
			}
			wait.Done()
		}(f)
	}
	wait.Wait()

	if *argWatch {
		watch(input, &r)
	}
}

// watch file modifications and call appropriate renderer actions
func watch(root string, renderer *renderer.Renderer) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				path := event.Name
				switch {
				case event.Op&fsnotify.Write == fsnotify.Write:
					// log.Println("Modified file: ", event.Name)
					if isTargetFile(path) {
						log.Println("detect change (file)", path)
						renderer.Render(path)
					}
				case event.Op&fsnotify.Create == fsnotify.Create:
					if isTargetFile(path) {
						log.Println("new file detected:", path)
						renderer.Render(path)
					} else if isDir(path) {
						log.Println("new directory detected:", path)
						watcher.Add(path)
					}
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					// TODO
				case event.Op&fsnotify.Rename == fsnotify.Rename:
					// TODO
				case event.Op&fsnotify.Chmod == fsnotify.Chmod:
					// TODO
				}
			case err := <-watcher.Errors:
				log.Println("error: ", err)
				done <- true
			}
		}
	}()

	for _, p := range getDirectories(root) {
		err = watcher.Add(p)
		if err != nil {
			log.Fatal(err)
		}
	}

	<-done
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

// get directories under root specified (recursively)
func getDirectories(root string) []string {
	var directories = []string{}

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			directories = append(directories, path)
		}
		return nil
	})

	return directories
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

// cleanse path string, for example, "some/path/" -> "some/path"
func cleanPath(path string) string {
	return filepath.Join(path)
}

// see if specified path is markdown file
func isTargetFile(path string) bool {
	if isDir(path) {
		return false
	}
	return filepath.Ext(path) == ".md"
}

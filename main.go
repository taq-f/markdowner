package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/go-fsnotify/fsnotify"
	zglob "github.com/mattn/go-zglob"
	"github.com/pkg/errors"
	"github.com/taq-f/markdowner/renderer"
	"github.com/taq-f/markdowner/style"
	"github.com/taq-f/markdowner/template"
)

func main() {
	argInputFile := flag.String("f", "", "input file/directory. if directory is specified, all markdown files under the directory will be rendered (recursively). if not specified, current directory will be target.")
	argOutDir := flag.String("o", "", "output directory. if not specified, output html file will be located in the same directory as input markdown file.")
	argImageInline := flag.Bool("i", false, "whether image files are embeded into html file. default: false.")
	argWatch := flag.Bool("w", false, "watch modification of markdown files and refresh html file as modification. default: false.")
	argCustomTemplate := flag.String("t", "", "custom html template file path.")
	argCustomStyle := flag.String("s", "", "custom stylesheet path")
	flag.Parse()

	log.Println("INFO : preparing...")

	// input file / directory
	// if not specified, current directory
	var input string
	if *argInputFile == "" {
		input, _ = filepath.Abs(filepath.Dir(os.Args[0]))
		log.Println("INFO : input file/directory not specified. current directory will be input directory:", input)
	} else {
		input = cleanPathStr(*argInputFile)
		log.Println("INFO : input file/directory set:", input)
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
		log.Println("INFO : output directory is not speficied. same as input:", outDir)
	} else {
		outDir = cleanPathStr(*argOutDir)

		log.Println("INFO : output directory:", outDir)
		log.Println("INFO : cleaning output directory")
		err := os.RemoveAll(outDir)
		if err != nil {
			log.Fatalln("ERROR: failed to clear output directory")
		}
		log.Println("INFO : cleaning successfully completed")
	}

	// custom styles
	customTemplate := *argCustomTemplate

	// custom styles
	var customStyles []string
	if *argCustomStyle == "" {
		customStyles = []string{}
	} else {
		customStyles = strings.Split(*argCustomStyle, " ")
	}

	r := renderer.Renderer{
		ImageInline: *argImageInline,
		Template:    template.Get(customTemplate),
		Style:       style.Get(&customStyles),
		OutDir:      outDir,
		BaseDir:     baseDir}

	files, err := getTargetFiles(input)
	if err != nil {
		log.Fatalln("ERROR: failed to find target files:", err)
	}

	log.Println("INFO : initialization completed")
	log.Printf("INFO : %d files detected", len(files))

	wait := new(sync.WaitGroup)
	wait.Add(len(files))

	var failed []string

	for _, f := range files {
		go func(file string) {
			err = r.Render(file)
			if err == nil {
				log.Printf("INFO : done: %s", file)
			} else {
				log.Printf("INFO : fail: %s: %s", file, err)
				failed = append(failed, file)
			}
			wait.Done()
		}(f)
	}
	wait.Wait()

	log.Printf("INFO : SUMMARY: all %d, success %d, fail %d", len(files), len(files)-len(failed), len(failed))

	if *argWatch {
		log.Println("INFO : start watching...")
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
		// when a file is modified (saved), "Write" occurs multiple times in
		// a very short time. to detect a event is originated from the same
		// operation, record time of events on the same file.
		modTimeTable := map[string]int64{}

		for {
			select {
			case event := <-watcher.Events:
				path := event.Name
				switch {
				case event.Op&fsnotify.Write == fsnotify.Write:
					if isTargetFile(path) {
						now := time.Now().UnixNano()
						t, exists := modTimeTable[path]
						var doRender bool
						if exists {
							span := now - t
							if span > (200 * 1000 * 1000) {
								doRender = true
								modTimeTable[path] = now
							}
						} else {
							modTimeTable[path] = now
							doRender = true
						}

						if doRender {
							log.Println("INFO : modification detected:", path)
							renderer.Render(path)
						}
					}
				case event.Op&fsnotify.Create == fsnotify.Create:
					if isTargetFile(path) {
						log.Println("INFO : new file detected:", path)
						renderer.Render(path)
					} else if isDir(path) {
						log.Println("INFO : new directory detected:", path)
						watcher.Add(path)
					}
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					if isDir(path) {
						watcher.Remove(path)
					}
				case event.Op&fsnotify.Rename == fsnotify.Rename:
					if isDir(path) {
						watcher.Remove(path)
					}
				case event.Op&fsnotify.Chmod == fsnotify.Chmod:
					// TODO
				}
			case err := <-watcher.Errors:
				log.Println("ERROR: watch error: ", err)
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
		return nil, errors.Wrapf(err, "failed to get all md files under %s", path)
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
func cleanPathStr(path string) string {
	return filepath.Join(path)
}

// see if specified path is markdown file
func isTargetFile(path string) bool {
	if isDir(path) {
		return false
	}
	return filepath.Ext(path) == ".md"
}
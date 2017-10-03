package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-fsnotify/fsnotify"
	zglob "github.com/mattn/go-zglob"
	"github.com/pkg/errors"
	exists "github.com/taq-f/go-exists"
	"github.com/taq-f/miniature-potato/renderer"
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

	// input file can also be specified without flag. command line arg without
	// flag is primary
	args := flag.Args()
	if len(args) >= 1 {
		argInputFile = &args[0]
	}

	// paths
	inputPath, basePath, outPath, err := getPath(*argInputFile, *argOutDir)

	if err != nil {
		log.Fatalf("ERROR: %v", err)
	}

	// custom template
	template := ""
	if *argCustomTemplate != "" {
		content, err := ioutil.ReadFile(*argCustomTemplate)
		if err != nil {
			log.Println("WARN : failed to read template. default template will be used:", err)
			content, _ = Asset("assets/template.html")
			template = string(content)
		} else {
			template = string(content)
		}
	} else {
		content, _ := Asset("assets/template.html")
		template = string(content)
	}

	// custom styles
	style := ""
	if *argCustomStyle != "" {
		content, err := ioutil.ReadFile(*argCustomStyle)
		if err != nil {
			log.Println("WARN : failed to read stylesheet. default style will be used:", err)
			content, _ = Asset("assets/default.css")
			style = string(content)
		} else {
			style = string(content)
		}
	} else {
		content, _ := Asset("assets/default.css")
		style = string(content)
	}
	style = "\n<style>\n" + style + "\n</style>\n"

	r := renderer.Renderer{
		ImageInline: *argImageInline,
		Template:    template,
		Style:       style,
		OutDir:      outPath,
		BaseDir:     basePath,
	}

	files, err := getTargetFiles(inputPath)
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
		watch(inputPath, &r)
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
	if !exists.File(path) {
		return nil, errors.New("file not found")
	}

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

// see if specified path is markdown file
func isTargetFile(path string) bool {
	if isDir(path) {
		return false
	}
	return filepath.Ext(path) == ".md"
}

// get directory paths
func getPath(input, output string) (inputPath, basePath, outPath string, err error) {
	// input path and base path
	if input == "" {
		// current directory if not specified
		// inputPath, _ = filepath.Abs(filepath.Dir(os.Args[0]))
		curPath, e := os.Getwd()
		if e != nil {
			err = e
			return
		}
		inputPath, e = filepath.Abs(curPath)
		if e != nil {
			err = e
			return
		}
		basePath = inputPath
	} else {
		// always handle path as absolute path
		i, e := filepath.Abs(input)
		if e != nil {
			err = e
			return
		}
		inputPath = i

		// input path must exist
		if !exists.File(inputPath) {
			err = fmt.Errorf("input path does not exists: %v", inputPath)
			return
		}

		if isDir(inputPath) {
			basePath = inputPath
		} else {
			basePath = filepath.Dir(inputPath)
		}
	}

	if output == "" {
		// same as input
		if isDir(inputPath) {
			outPath = inputPath
		} else {
			outPath = filepath.Dir(inputPath)
		}
	} else {
		outPath, _ = filepath.Abs(output)
	}

	return
}

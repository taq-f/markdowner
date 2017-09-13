package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/russross/blackfriday"
	"github.com/taq-f/markdowner/style"
	"github.com/taq-f/markdowner/template"
)

func main() {
	var outDir = flag.String("o", "default value", "help text")
	var inputFile = flag.String("f", "default value", "help text")
	flag.Parse()

	data, err := ioutil.ReadFile(*inputFile)
	if err != nil {
		fmt.Println("ERROR IN READING FILE", err)
		return
	}

	template := template.Get()
	style := style.Get()
	markdowned := blackfriday.MarkdownBasic(data)

	output := template
	output = strings.Replace(output, "{{{content}}}", string(markdowned), -1)
	output = strings.Replace(output, "{{{style}}}", style, -1)

	ioutil.WriteFile(*outDir, []byte(output), os.ModeAppend)
}

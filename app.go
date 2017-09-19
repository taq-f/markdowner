package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	zglob "github.com/mattn/go-zglob"
	"github.com/russross/blackfriday"
	"github.com/sourcegraph/syntaxhighlight"
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
		err = render(template, style, f, *argOutDir, *argInputFile)
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

func render(template string, style string, input string, outDir string, baseDir string) error {
	data, err := ioutil.ReadFile(input)
	if err != nil {
		fmt.Println("ERROR IN READING FILE", err)
		return err
	}

	markdowned := string(blackfriday.MarkdownCommon(data))
	markdowned = highlightCode(markdowned, filepath.Dir(input))

	output := template
	output = strings.Replace(output, "{{{content}}}", markdowned, -1)
	output = strings.Replace(output, "{{{style}}}", style, -1)

	out, err := getOutPath(input, outDir, baseDir)

	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(out), os.ModeDir)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(out, []byte(output), os.ModeAppend)
	if err != nil {
		return err
	}

	return nil
}

func highlightCode(html string, inputPath string) string {

	reader := strings.NewReader(html)
	doc, _ := goquery.NewDocumentFromReader(reader)

	doc.Find("code[class*=\"language-\"]").Each(func(i int, s *goquery.Selection) {
		oldCode := s.Text()
		formatted, err := syntaxhighlight.AsHTML([]byte(oldCode))
		if err != nil {
			log.Fatal(err)
		}
		s.SetHtml(string(formatted))
	})

	// 画像をbase64で含める場合
	// doc.Find("img").Each(func(i int, s *goquery.Selection) {
	// 	src, _ := s.Attr("src")
	// 	path := filepath.Join(inputPath, src)
	// 	mime := mime.TypeByExtension(filepath.Ext(path))
	// 	base64, _ := imageToBase64(path)
	// 	srcEnced := fmt.Sprintf("data:%s;base64,%s", mime, base64)

	// 	s.SetAttr("src", srcEnced)
	// })

	new, _ := doc.Html()

	new = strings.Replace(new, "<html><head></head><body>", "", 1)
	new = strings.Replace(new, "</body></html>", "", 1)

	return new
}

func getOutPath(input string, specified string, baseDir string) (string, error) {
	if specified == "" {
		// same directory as input
		extension := filepath.Ext(input)
		fileBasename := input[:utf8.RuneCountInString(input)-utf8.RuneCountInString(extension)]
		newFilename := fileBasename + ".html"

		return newFilename, nil
	}

	out := filepath.Join(specified, input[utf8.RuneCountInString(baseDir):])
	extension := filepath.Ext(out)
	fileBasename := out[:utf8.RuneCountInString(out)-utf8.RuneCountInString(extension)]
	out = fileBasename + ".html"

	return out, nil
}

func imageToBase64(path string) (string, error) {
	imageFile, err := os.Open(path)
	if err != nil {
		return "", err
	}

	image, err := ioutil.ReadAll(imageFile)

	if err != nil {
		return "", err
	}

	imgEnc := base64.StdEncoding.EncodeToString(image)
	return imgEnc, nil
}

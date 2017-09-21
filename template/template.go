package template

import (
	"io/ioutil"
	"log"
)

var defaultTemplate = `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-type" content="text/html;charset=UTF-8">
{{{style}}}
</head>
<body>
{{{content}}}
</body>
</html>
`

// Get html template string
func Get(custom string) string {
	if custom == "" {
		return defaultTemplate
	}

	content, err := ioutil.ReadFile(custom)
	if err != nil {
		log.Println("WARN : failed to read template:", err)
		return ""
	}

	return string(content)
}

package template

func Get() string {
	return `<!DOCTYPE html>
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
}

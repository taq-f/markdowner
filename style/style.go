package style

import (
	"io/ioutil"
	"log"
	"strings"
)

var template = `
<style>
{{{content}}}
</style>
`

var defaultStyle = `
/*! Color themes for Google Code Prettify | MIT License | github.com/jmblog/color-themes-for-google-code-prettify */
/* pre{background:#fff;font-family:Menlo,Bitstream Vera Sans Mono,DejaVu Sans Mono,Monaco,Consolas,monospace;border:0!important} */ .pln{color:#333}ol.linenums{margin-top:0;margin-bottom:0;color:#ccc}li.L0,li.L1,li.L2,li.L3,li.L4,li.L5,li.L6,li.L7,li.L8,li.L9{padding-left:1em;background-color:#fff;list-style-type:decimal}@media screen{.str{color:#183691}.kwd{color:#a71d5d}.com{color:#969896}.typ{color:#0086b3}.lit{color:#0086b3}.pun{color:#333}.opn{color:#333}.clo{color:#333}.tag{color:navy}.atn{color:#795da3}.atv{color:#183691}.dec{color:#333}.var{color:teal}.fun{color:#900}}

body {
	color: rgba(0,0,0,.87);
	font-family: "Segoe WPC", "Segoe UI", "SFUIText-Light", "HelveticaNeue-Light", sans-serif, "Droid Sans Fallback";
	font-size: 14px;
	line-height: 22px;
	margin: 0 auto;
	padding: 0 12px;
	width: 700px;
	word-wrap: break-word;
}

img {
	max-width: 100%;
	max-height: 100%;
	display: block;
}

a {
	color: #4080D0;
	text-decoration: none;
}
a:focus {
	outline: 1px solid -webkit-focus-ring-color;
	outline-offset: -1px;
}
a:hover {
	color: #4080D0;
	text-decoration: underline;
}

table {
	border-collapse: collapse;
}
table > thead > tr > th {
	text-align: left;
	border-bottom: 1px solid;
}
table > thead > tr > th,
table > thead > tr > td,
table > tbody > tr > th,
table > tbody > tr > td {
	padding: 5px 10px;
}
table > tbody > tr + tr > td {
	border-top: 1px solid;
}

blockquote {
	background: rgba(127, 127, 127, 0.1);
	border-color: rgba(0, 122, 204, 0.5);
	border-left: 5px solid;
	margin: 0 7px 0 5px;
	padding: 0 16px 0 10px;
}

code {
	font-family: Menlo, Monaco, Consolas, "Droid Sans Mono", "Courier New", monospace, "Droid Sans Fallback";
	font-size: 14px;
	line-height: 19px;
}

pre {
	background-color: #f8f8f8;
	border: 1px solid #cccccc;
	border-radius: 3px;
	overflow-x: auto;
	white-space: pre-wrap;
	overflow-wrap: break-word;

	padding: 16px;
	border-radius: 3px;
	overflow: auto;
}

/*
pre:not(.hljs),
pre.hljs code > div {
	padding: 16px;
	border-radius: 3px;
	overflow: auto;
}
*/

hr {
	border: 0;
	height: 2px;
	border-bottom: 2px solid;
}

h1 {
	line-height: 1.2;
	padding-bottom: 0.3em;
	text-align: center;
	border-bottom: 0;
}

h1, h2, h3 {
	font-weight: normal;
}

h2 {
	border-bottom: 2px solid #d4d4d4;
	margin: 25px 0;
	padding-bottom: 10px;
}

h3 {
	font-size: 1.4em;
	margin: 25px 0;
}

h1 code,
h2 code,
h3 code,
h4 code,
h5 code,
h6 code {
	font-size: inherit;
	line-height: auto;
}

body > p, table, blockquote, pre {
	margin-left: 10px;
	margin-right: 10px;
}

/*
:not(pre):not(.hljs) > code {
	color: #A31515;
	font-size: inherit;
	font-family: inherit;
}
*/

li {
	padding-top: 5px;
	padding-bottom: 5px;
}
`

// Get style tag string to embed
func Get(custom *[]string) string {
	if custom == nil || len(*custom) == 0 {
		return strings.Replace(template, "{{{content}}}", defaultStyle, 1)
	}

	var style string

	for _, path := range *custom {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			log.Println("WARNING: style sheet read error: ", err)
			continue
		}

		style += string(content) + "\n"
	}

	return strings.Replace(template, "{{{content}}}", style, 1)
}

/* Copyright (c) 2014-2015, Daniel Martí <mvdan@mvdan.cc> */
/* See LICENSE for licensing information */

package main

import "html/template"

var tmpl *template.Template

func loadTemplates() {
	for name, s := range templates {
		var t *template.Template
		if tmpl == nil {
			tmpl = template.New(name)
		}
		if name == tmpl.Name() {
			t = tmpl
		} else {
			t = tmpl.New(name)
		}
		if _, err := t.Parse(s); err != nil {
			panic("could not load templates")
		}
	}
}

var templates = map[string]string{
	"/": `<html>
<body style="text-align:center">
<div style="inline-block">
	<form action="{{.SiteURL}}" method="post" enctype="multipart/form-data">
		<textarea cols=80 rows=24 name="{{.FieldCode}}"></textarea>
		<br/>
		<button type="submit">Upload Go code</button>
	</form>
	<br/>
	<form action="{{.SiteURL}}" method="post" enctype="multipart/form-data">
		<input type="file" name="{{.FieldCode}}"></input>
		<button type="submit">Upload file</button>
	</form>
</div>
</body>
</html>
`,
}

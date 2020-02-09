package web

import (
	"fmt"
	"github.com/rakyll/statik/fs"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
)

type Templates struct {
	fs   http.FileSystem
	base *template.Template
}

func NewTemplates() (*Templates, error) {
	f, err := fs.New()
	if err != nil {
		return nil, err
	}

	tpl := &Templates{fs: f}

	src, err := tpl.readFile("/templates/index.html")
	if err != nil {
		return nil, err
	}

	tpl.base, err = template.New("html").Parse(src)
	if err != nil {
		return nil, err
	}

	customFuncs := template.FuncMap{
		"indent": func(spaces int, v string) string {
			pad := strings.Repeat(" ", spaces)
			return pad + strings.Replace(v, "\n", "\n"+pad, -1)
		},
		"inc": func(v int) string {
			return fmt.Sprintf("%v", v+1)
		},
		"status": func(v string) string {
			switch v {
			case "SUCCESS":
				return "\u2705"
			case "FAILURE":
				return "\u274c"
			case "ENQUEUED":
				return "\u23f3"
			case "EXECUTING":
				return "\U0001f3c3\u200d\u2640\ufe0f"
			case "FINISHED":
				return "\U0001f3c1"
			default:
				return v
			}
		},
		"githubStatus": func(v string) string {
			switch v {
			case "success":
				return ":heavy_check_mark:"
			case "failure":
				return ":x:"
			default:
				return v
			}
		},
	}

	tpl.base = tpl.base.Funcs(customFuncs)

	return tpl, nil
}

func (t *Templates) readFile(filename string) (string, error) {
	// Access individual files by their paths.
	r, err := t.fs.Open(filename)
	if err != nil {
		return "", err
	}
	//noinspection GoUnhandledErrorResult
	defer r.Close()
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

func (t *Templates) New(name string) (*template.Template, error) {
	tpl, err := t.base.Clone()
	if err != nil {
		return nil, err
	}
	md, err := t.readFile(fmt.Sprintf("/templates/%s.md", name))
	if err != nil {
		return nil, err
	}
	_, err = tpl.New("markdown").Parse(md)
	if err != nil {
		return nil, err
	}
	return tpl, nil
}

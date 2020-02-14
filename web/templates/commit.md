{{define "title"}}{{slice .Commit 0 7}} @ {{.Login}}/{{.Repo}}{{end -}}
# Commit Report

{{.Status | status}} [{{slice .Commit 0 7}}](https://github.com/{{.Login}}/{{.Repo}}/commit/{{.Commit}}) from [{{.Login}}/{{.Repo}}](https://github.com/{{.Login}}/{{.Repo}})

{{if .Checks}}
## Checks

{{range .Checks -}}
* {{.Status | status}} `{{.Name}}`
  {{- if .Output}}
  ```text
{{.Output | indent 2 | unescape -}}
  ```
  {{- end}}
{{end -}}
{{end}}

* [Back to main page](..)

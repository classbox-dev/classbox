{{define "title"}}{{slice .Commit 0 7}} @ {{.Login}}/{{.Repo}}{{end -}}
# Commit Report

[{{.Status}}] [{{slice .Commit 0 7}}](https://github.com/{{.Login}}/{{.Repo}}/commit/{{.Commit}}) from [{{.Login}}/{{.Repo}}](https://github.com/{{.Login}}/{{.Repo}})

{{if .Checks}}
## Checks

{{range .Checks -}}
* [{{.Status}}] `{{.Name}}`
  {{- if .Output}}
  ```text
{{.Output | indent 2 -}}```
  {{- end}}
{{end -}}
{{end}}
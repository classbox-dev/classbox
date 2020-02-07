# Commit Report

[{{.Status}}] [{{slice .Commit 0 7}}](https://github.com/{{.Login}}/{{.Repo}}/commit/{{.Commit}}) from [{{.Login}}/{{.Repo}}](https://github.com/{{.Login}}/{{.Repo}})

## Checks

{{range .Checks -}}
* [{{.Status}}] `{{.Name}}`
  {{- if .Output}}
  ```text
{{.Output | indent 2 -}}```
  {{- end}}
{{end -}}

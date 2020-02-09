{{range . -}}
* {{.Status | githubStatus}} `{{.Name}}`
  {{- if .Output}}
  ```text
{{.Output | indent 2 }}
  ```
  {{- end}}
{{end -}}

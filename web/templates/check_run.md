Only non-cached or failed checks are shown here. See the [full report]({{ .Url }}) for more details.

{{range .Stages -}}
{{ if or (not .Cached) (ne .Status "success") -}}
* {{.Status | githubStatus}} `{{.Name}}`
  {{- if .Output}}
  ```text
{{.Output | indent 2 | unescape }}
  ```
  {{- end}}
{{- end}}
{{end -}}

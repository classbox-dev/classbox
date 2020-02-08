{{define "title"}}stdlib @ hsecode{{end -}}
# stdlib

## Scoreboard

| Login | Count | Score |
|-------|-------------|-------|
{{range .Stats -}}
| [{{ .Login }}](https://github.com/{{.Login}}) | {{ .Count }} |  {{ .Score }} |
{{end -}}


## Tests

| ID | Description | Score |
|-------|-------------|-------|
{{range .Tests -}}
| `{{ .Name }}` | {{ .Description }} |  {{ .Score }} |
{{end -}}

{{define "title"}}stdlib @ hsecode{{end -}}
# stdlib
Hi, {{ .User.Login }}! | [Logout](logout)

Your working repository: [{{ .User.Login }}/{{ .User.Repo }}](https://github.com/{{ .User.Login }}/{{ .User.Repo }})

## Documentaion

* [Honour code](signin?step=honour_code)
* [Prerequisites](prerequisites)
* [Quickstart](quickstart)
* [stdlib documentation]({{.DocsURL}})

## Tests

* Total score: {{.Stats.Score}} out of {{.Stats.Total}}
* Grade (*Theory of Algorithms* only): **{{.Grade}} out of 10**
* [Grading policy](grading)
* [Scoreboard](scoreboard)

| ID | Description | Score | Passed |
|----|-------------|-------|--------|
{{range .Stats.Tests -}}
| `{{ .Name }}` | {{ .Description }} |  {{ .Score }} | {{if .Passed }}✅{{else}}⬜️{{end}} |
{{end -}}

{{define "title"}}stdlib @ hsecode{{end -}}
# stdlib

{{if not .User -}}
stdlib is a challenge to implement a library of data structures and algorithms in Golang using only documentation.

[Sign in](signin) via GitHub to go down the rabbit hole.

{{- else -}}

Hi, {{ .User.Login }}! | [Logout](logout)

Your working repository: [{{ .User.Login }}/{{ .User.Repo }}](https://github.com/{{ .User.Login }}/{{ .User.Repo }})

## Quickstart

* Read [stdlib documentation]({{.DocsURL}}).
* Start implementing some part of the library to pass one of the tests below
* Commit and push the implementation to your working repository.
* The code is automatically tested on each commit. You can see the results on the [GitHub commits page](https://github.com/{{ .User.Login }}/{{ .User.Repo }}/master), your feed, or the scoreboard below.
{{- end}}

## Scoreboard

| # | Login | Passed tests | Score |
|---|-------|--------------|-------|
{{range $i, $x := .Stats -}}
| {{$i | inc}} | [{{ $x.Login }}](https://github.com/{{$x.Login}}) | {{ $x.Count }} | {{ $x.Score }} |
{{end -}}


## Tests

| ID | Description | Score |
|-------|-------------|-------|
{{range .Tests -}}
| `{{ .Name }}` | {{ .Description }} |  {{ .Score }} |
{{end -}}

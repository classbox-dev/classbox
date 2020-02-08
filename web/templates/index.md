{{define "title"}}stdlib @ hsecode{{end -}}
# stdlib

stdlib is a challenge to implement a library of data structures and algorithms in Golang using only documentation.

## Quickstart

1. [Sign up](signup) with your Github account. Working private repository will be created automatically.
2. Read [stdlib documentation]({{.DocsURL}})

## Scoreboard

| # | Login | Passed | Score |
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

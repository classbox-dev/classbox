{{define "title"}}Scoreboard @ hsecode{{end -}}
# Scoreboard
{{if not .User -}}
[Sign in](../signin) to see the scoreboard.
{{- else -}}
| # | Login | Passed tests | Score |
|---|-------|--------------|-------|
{{range $i, $x := .Stats -}}
| {{$i | inc}} | [{{ $x.Login }}](https://github.com/{{$x.Login}}) | {{ $x.Count }} | {{ $x.Score }} |
{{end -}}
{{end}}

* [Back to main page](..)

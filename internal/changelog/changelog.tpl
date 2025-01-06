# Changelog

## {{ .Version }} - {{ .Date }}{{"\n"}}

{{- if eq (len .Commits) 0 }}
- No changes
{{ end}}

{{- if (index .Commits 0) }}
### BREAKING CHANGES
{{- range (index .Commits 0) }}
- {{ if .Scope -}} **{{ .Scope }}:** {{- else -}} {{ .Type }}: {{- end }} {{ .Message }} ({{ .Sha.Short }})
{{- end }}
{{ end }}

{{- if (index .Commits 1) }}
### Features
{{- range (index .Commits 1) }}
- {{ if .Scope -}} **{{ .Scope }}:** {{ end -}} {{ .Message }} ({{ .Sha.Short }})
{{- end }}
{{ end }}

{{- if (index .Commits 2) }}
### Bug Fixes
{{- range (index .Commits 2) }}
- {{ if .Scope -}} **{{ .Scope }}:** {{ end -}} {{ .Message }} ({{ .Sha.Short }})
{{- end }}
{{ end }}

{{- if (index .Commits 3) }}
### Miscellaneous
{{- range (index .Commits 3) }}
- {{ if .Scope -}} **{{ .Scope }}:** {{- else -}} {{ .Type }}: {{- end }} {{ .Message }} ({{ .Sha.Short }})
{{- end }}
{{ end }}
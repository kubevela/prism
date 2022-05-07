{{- define "vela-prism.labels" -}}
app.kubernetes.io/name: vela-prism
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "vela-prism.selector-labels" -}}
app: vela-prism
{{- end -}}

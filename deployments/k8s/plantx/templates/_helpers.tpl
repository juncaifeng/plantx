{{/*
Expand the name of the chart.
*/}}
{{- define "plantx.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "plantx.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "plantx.labels" -}}
helm.sh/chart: {{ include "plantx.chart" . }}
app.kubernetes.io/name: {{ include "plantx.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Expand the name of the chart.
*/}}
{{- define "cita-cloud-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}


{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "cita-cloud-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "cita-cloud-operator.labels" -}}
helm.sh/chart: {{ include "cita-cloud-operator.chart" . }}
{{ include "cita-cloud-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "cita-cloud-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "cita-cloud-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "azuero-relief.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- define "azuero-relief.fullname" -}}
{{- default (printf "%s-%s" .Release.Name (include "azuero-relief.name" .)) .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- define "azuero-relief.selectorLabels" -}}
app.kubernetes.io/name: {{ include "azuero-relief.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
{{- define "azuero-relief.labels" -}}
{{ include "azuero-relief.selectorLabels" . }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | quote }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}
{{- define "azuero-relief.serviceAccount" -}}
{{- if .Values.serviceAccount.create -}}
{{- default (include "azuero-relief.fullname" .) .Values.serviceAccount.name -}}
{{- else -}}
{{- default "default" .Values.serviceAccount.name -}}
{{- end -}}
{{- end -}}
{{- define "azuero-relief.image" -}}
{{- if .Values.image.digest -}}
{{- if not (regexMatch "^sha256:[a-f0-9]{64}$" .Values.image.digest) -}}
{{- fail "image.digest must be a sha256 digest" -}}
{{- end -}}
{{- printf "%s@%s" .Values.image.repository .Values.image.digest -}}
{{- else -}}
{{- printf "%s:%s" .Values.image.repository (required "Set image.tag or image.digest to a published image" .Values.image.tag) -}}
{{- end -}}
{{- end -}}

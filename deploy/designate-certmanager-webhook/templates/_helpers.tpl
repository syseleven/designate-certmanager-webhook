{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "designate-certmanager-webhook.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "designate-certmanager-webhook.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "designate-certmanager-webhook.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "designate-certmanager-webhook.selfSignedIssuer" -}}
{{ printf "%s-selfsign" (include "designate-certmanager-webhook.fullname" .) }}
{{- end -}}

{{- define "designate-certmanager-webhook.rootCAIssuer" -}}
{{ printf "%s-ca" (include "designate-certmanager-webhook.fullname" .) }}
{{- end -}}

{{- define "designate-certmanager-webhook.rootCACertificate" -}}
{{ printf "%s-ca" (include "designate-certmanager-webhook.fullname" .) }}
{{- end -}}

{{- define "designate-certmanager-webhook.servingCertificate" -}}
{{ printf "%s-webhook-tls" (include "designate-certmanager-webhook.fullname" .) }}
{{- end -}}

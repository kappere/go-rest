apiVersion: v1
kind: ConfigMap
metadata:
  name: {{.appname}}
  labels:
    app: {{.appname}}
  namespace: default
data:
  {{.appname}}.yaml: |
    K8S_CONFIGMAP: {{.appname}}.yaml
---
apiVersion: v1
kind: Secret
metadata:
  name: {{.appname}}
  labels:
    app: {{.appname}}
  namespace: default
type: Opaque
stringData:
  {{.appname}}-secret.yaml: |
    K8S_SECRET: {{.appname}}-secret.yaml
apiVersion: v1
kind: Service
metadata:
  creationTimestamp: null
  labels:
    app: {{.appname}}
  name: {{.appname}}
  namespace: default
spec:
  type: ClusterIP
  ports:
  - name: http
    port: 80
  selector:
    app: {{.appname}}
---
apiVersion: apps/v1
kind: Deployment
metadata:
  creationTimestamp: null
  labels:
    app: {{.appname}}
  name: {{.appname}}
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{.appname}}
  strategy: {}
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: {{.appname}}
    spec:
      containers:
      - name: {{.appname}}
        image: {{.appname}}:1.0.0
        ports:
        - containerPort: 80
        env:
          - name: PROFILE
            value: prod
        volumeMounts:
        - name: config-volume
          mountPath: /etc/{{.appname}}
        - name: log-volume
          mountPath: /var/log/{{.appname}}
      volumes:
      - name: config-volume
        configMap:
          name: {{.appname}}
      - name: log-volume
        hostPath:
          path: /var/log/{{.appname}}

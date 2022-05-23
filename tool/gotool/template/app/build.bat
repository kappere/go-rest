go mod vendor
minikube image build -t {{.appname}}:1.0.0 .
@REM docker build -t {{.appname}}:1.0.0 .
rd /s/q vendor
go mod vendor
minikube image build -t {{.appname}}:1.0.0 .
minikube image save {{.appname}}:1.0.0 {{.appname}}-1.0.0.tar
@REM docker build -t {{.appname}}:1.0.0 .
rd /s/q vendor
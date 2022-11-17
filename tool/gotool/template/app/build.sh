go mod vendor
minikube image build -t {{.appname}}:1.0.0 .
minikube image save {{.appname}}:1.0.0 {{.appname}}-1.0.0.tar
# docker build -t {{.appname}}:latest .
rm -rf vendor
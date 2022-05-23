go mod vendor
minikube image build -t {{.appname}}:1.0.0 .
# docker build -t {{.appname}}:latest .
rm -rf vendor
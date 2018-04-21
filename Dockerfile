FROM golang:latest

MAINTAINER AmitD <amit.das@openebs.io>

# Install kubectl
ENV KUBE_LATEST_VERSION="v1.9.3"

RUN curl -L https://storage.googleapis.com/kubernetes-release/release/${KUBE_LATEST_VERSION}/bin/linux/amd64/kubectl -o /usr/local/bin/kubectl \
 && chmod +x /usr/local/bin/kubectl \
 && kubectl version --client

# Install go tools
RUN go get github.com/DATA-DOG/godog/cmd/godog
RUN go get -u github.com/golang/dep/cmd/dep

# Add source code
RUN mkdir -p /go/src/github.com/AmitKumarDas/litmus
ADD . /go/src/github.com/AmitKumarDas/litmus/
WORKDIR /go/src/github.com/AmitKumarDas/litmus/

# Go dep
RUN dep ensure

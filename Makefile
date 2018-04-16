# list only our namespaced directories
PACKAGES = $(shell go list ./... | grep -v '/vendor/')

.PHONY: all
all: format metalint compile

# `make setup` needs to be run in a completely new environment
# In case of go related issues, run below commands & verify:
# go version    # ensure go1.9.1 or above
# go env        # ensure if GOPATH is set
# echo $PATH    # ensure if $GOPATH/bin is set
.PHONY: setup
setup:
	@echo "------------------"
	@echo "--> Running setup"
	@echo "------------------"
	@go get -u -v github.com/golang/lint/golint
	@go get -u -v golang.org/x/tools/cmd/goimports
	@go get -u -v github.com/golang/dep/cmd/dep
	@go get -u -v github.com/DATA-DOG/godog/cmd/godog
	@go get -u -v github.com/alecthomas/gometalinter
	@gometalinter --install

.PHONY: format
format:
	@echo "------------------"
	@echo "--> Running go fmt"
	@echo "------------------"
	@go fmt $(PACKAGES)

.PHONY: lint
lint:
	@echo "------------------"
	@echo "--> Running golint"
	@echo "------------------"
	@golint $(PACKAGES)
	@echo "------------------"
	@echo "--> Running go vet"
	@echo "------------------"
	@go vet $(PACKAGES)

.PHONY: metalint
metalint:
	@echo "------------------"
	@echo "--> Running metalinter"
	@echo "------------------"
	@gometalinter $(PACKAGES)

.PHONY: compile
compile:
	@echo "------------------"
	@echo "--> Check compilation"
	@echo "------------------"
	@go test $(PACKAGES)

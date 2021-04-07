BINARY = DynDNS
GOARCH = amd64

VERSION?=latest

# Symlink into GOPATH
GITHUB_USERNAME=praveenprem
BUILD_DIR=${GOPATH}/src/github.com/${GITHUB_USERNAME}/${BINARY}
BIN_DIR=${BUILD_DIR}/bin
CURRENT_DIR=\$(shell pwd)
BUILD_DIR_LINK=\$(shell readlink ${BUILD_DIR})

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS = -ldflags "-X main.VERSION=${VERSION}"

export GO111MODULE=on
export GOPROXY=direct
export GOSUMDB=off

help:
	@echo "Usage: make [command] [option]"
	@echo "\nCommands:"
	@echo "\t run \t\t-- Run dev environment on machine"
	@echo "\t linux \t\t-- Build application for Linux operating systems"
	@echo "\t build \t\t-- Build the application for Darwin"
	@echo "\t package \t-- Tar installation archive of binary and resources"
	@echo "\t release \t-- Release new version of the application (build, package and clean)"
	@echo "\t install \t-- Install the binary to system"
	@echo "\t upgrade \t-- Upgrade Go Mod versions"
	@echo "\t dep \t\t-- Resolve mods in the go.mod file"
	@echo "\t clean \t\t-- Clean build artefacts post package"
	@echo "\nOptions:"
	@echo "\t BINARY \t-- Name of the application"
	@echo "\t GOARCH \t-- Architecture of the application"
	@echo "\t VERSION \t-- Application version"

run:
	@go run .

linux:
	GOOS=linux GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BIN_DIR}/${BINARY}-linux-${GOARCH}/${VERSION}/${BINARY}/${BINARY} .

build:
	GOOS=darwin GOARCH=${GOARCH} go build ${LDFLAGS} -o ${BIN_DIR}/${BINARY} .

package:
	@cp -r resources/ ${BIN_DIR}/${BINARY}-linux-${GOARCH}/${VERSION}/${BINARY}/
	@mkdir -p ${BIN_DIR}/${VERSION}/
	@tar -cvjf ${BIN_DIR}/${VERSION}/${BINARY}-linux-${GOARCH}.tar -C ${BIN_DIR}/${BINARY}-linux-${GOARCH}/${VERSION}/ .

release: linux package clean

install:
	install ${BIN_DIR}/${BINARY} /usr/local/bin/${BINARY}

upgrade:
	@go get -u ./...

dep:
	@go list -m -u all

clean:
	@rm -rf ${BIN_DIR}/${BINARY}-*-${GOARCH}
	@rm -rf ${BIN_DIR}/${BINARY}

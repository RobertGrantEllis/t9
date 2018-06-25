# This makefile requires the go tools and dep, which can be installed on a Mac with Homebrew:
# brew install go dep
# This makefile also requires docker and docker-compose (https://www.docker.com/)
# See also subordinate makefile dependencies

# Parameters
BINARY_NAME=$(shell basename `pwd`)
LINUX_BINARY_NAME=$(BINARY_NAME)-linux-amd64

# Commands
GOCMD=go
GOFMT=$(GOCMD) fmt
GOINSTALL=$(GOCMD) install
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test

DEP=dep
MAKE=make
RM=rm
DOCKERCOMPOSE=docker-compose

all: generate vendor test install
generate:
	# the following makefiles have their own dependencies. See comments in each one.
	$(MAKE) -C frontend
	rm -rf bindata/frontend && cp -R frontend/dist bindata/frontend
	$(MAKE) -C bindata
	$(MAKE) -C proto
	$(MAKE) -C mocks
	$(GOFMT) ./...
vendor:
	$(DEP) ensure
test: 
	$(GOTEST) -v ./...
install:
	$(GOINSTALL) -v
clean: 
	$(GOCLEAN) -v
	$(RM) -rf $(BINARY_LINUX) vendor output
	$(MAKE) -C frontend clean
	$(MAKE) -C bindata clean
	$(MAKE) -C proto clean
	$(MAKE) -C mocks clean

linux: generate vendor test
	GOOS=linux GOARCH=amd64 go build -o "$(LINUX_BINARY_NAME)" .
	@echo "see binary $(LINUX_BINARY_NAME)"

docker: generate vendor test
	$(DOCKERCOMPOSE) build
# Makefile for the calcula-aposentadoria project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GORUN=$(GOCMD) run
BINARY_NAME=calcula-aposentadoria

all: build

build:
	$(GOBUILD) -o $(BINARY_NAME) main.go

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

run:
	$(GORUN) main.go

test:
	$(GOTEST)

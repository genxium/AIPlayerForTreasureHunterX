PROJECTNAME=server
ROOT_DIR=$(shell pwd)
all: help

run: build
	ServerEnv=TEST $(ROOT_DIR)/AI


build:
	go build -o $(ROOT_DIR)/AI

.PHONY: help

help: Makefile
	@echo
	@echo " Choose a command run:"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo
 

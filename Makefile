VERSION=$(shell git rev-parse --short HEAD)
BUILD_PATH := $(realpath ./build/main)

define with_env
	$(eval include $(1))
	$(eval export)
endef

chmod-deploy-scripts:
	chmod +x ./deploy/start.sh

executable-path:
	@echo $(BUILD_PATH)

dev:
	$(call with_env,./.env)
	go run ./main.go

build-main:
	mkdir -p build
	CGO_ENABLED=0 go build -o ./build/main -v ./main.go

chmod-main:
	chmod +x ./build/main

build: build-main chmod-main

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
	CGO_ENABLED=1 go build -o ./build/main -v ./main.go

chmod-main:
	chmod +x ./build/main

build: build-main chmod-main

# Frontend targets
frontend-install:
	cd frontend && npm install

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

dev-with-frontend:
	@echo "Starting Go backend on :9999 and Vite dev server on :5173"
	@echo "Access app at http://127.0.0.1:5173"
	$(call with_env,./.env)
	@trap 'kill 0' EXIT; \
	go run ./main.go & \
	cd frontend && npm run dev

build-all: frontend-build build

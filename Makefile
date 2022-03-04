BUILD_PATH := $(realpath ./build/localserver)

chmod-deploy-scripts:
	chmod +x ./deploy/start.sh

executable-path:
	@echo $(BUILD_PATH)

dev:
	with-dotenv .env go run ./cmd/localserver/*.go

build-localserver:
	mkdir -p build
	CGO_ENABLED=0 go build -o ./build/localserver -v ./cmd/localserver/*.go

chmod-localserver:
	chmod +x ./build/localserver

build: build-localserver chmod-localserver

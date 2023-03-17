# Define
VERSION=0.1.2
BUILD=$(shell git rev-parse HEAD)

.PHONY: image
image:
	docker build $(CURDIR) \
		--file build/Dockerfile.local \
		--build-arg "version=$(VERSION)" \
		--build-arg "commit=$(BUILD)" \
		--file build/Dockerfile.local \
		--tag registry.deflinhec.dev/nakama:latest

publish:
	docker push registry.deflinhec.dev/nakama:latest

default: image




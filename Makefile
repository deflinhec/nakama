# Define
VERSION=0.6.0
PROJECT=casino-383511
COMMIT=$(shell git rev-parse HEAD)
BUILD=$(shell git rev-parse --short HEAD)

.PHONY: image
image:
	docker build $(CURDIR) \
		--file build/Dockerfile.local \
		--build-arg "version=$(VERSION)" \
		--build-arg "commit=$(COMMIT)" \
		--tag gcr.io/$(PROJECT)/nakama:v$(VERSION)-$(BUILD)
	docker tag gcr.io/$(PROJECT)/nakama:v$(VERSION)-$(BUILD) \
		gcr.io/$(PROJECT)/nakama:develop

.PHONY: pluginbuilder-image
pluginbuilder-image:
	docker build $(CURDIR) \
		--file build/pluginbuilder/Dockerfile \
		--build-arg "version=$(VERSION)" \
		--build-arg "commit=$(COMMIT)" \
		--tag gcr.io/$(PROJECT)/nakama/pluginbuilder:latest

.PHONY: generate
generate:
	go generate -x ./... && \
	(cd console/ui && npm clean-install && npm run-script build)

default: image plugin-builder-image




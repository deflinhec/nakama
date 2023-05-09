# Define
VERSION=0.4.1
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

.PHONY: upgrade
upgrade: export GOPRIVATE=github.com/bcasino
upgrade: 
	go get github.com/bcasino/nakama-web && \
	go get github.com/bcasino/nakama-api && \
	go mod tidy && go mod vendor

default: image plugin-builder-image




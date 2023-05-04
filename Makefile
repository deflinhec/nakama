# Define
VERSION=0.2.10
PROJECT=casino-383511
BUILD=$(shell git rev-parse HEAD)

.PHONY: image
image:
	docker build $(CURDIR) \
		--file build/Dockerfile.local \
		--build-arg "version=$(VERSION)" \
		--build-arg "commit=$(BUILD)" \
		--tag gcr.io/$(PROJECT)/nakama:v$(VERSION)
	docker tag gcr.io/$(PROJECT)/nakama:v$(VERSION) \
		gcr.io/$(PROJECT)/nakama:latest

.PHONY: pluginbuilder-image
pluginbuilder-image:
	docker build $(CURDIR) \
		--file build/pluginbuilder/Dockerfile \
		--build-arg "version=$(VERSION)" \
		--build-arg "commit=$(BUILD)" \
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




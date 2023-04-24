# Define
VERSION=0.2.8
BUILD=$(shell git rev-parse HEAD)

.PHONY: image
image:
	docker build $(CURDIR) \
		--file build/Dockerfile.local \
		--build-arg "version=$(VERSION)" \
		--build-arg "commit=$(BUILD)" \
		--tag registry.deflinhec.dev/nakama:v$(VERSION)
	docker tag registry.deflinhec.dev/nakama:v$(VERSION) \
		registry.deflinhec.dev/nakama:latest

.PHONY: publish
publish:
	docker push registry.deflinhec.dev/nakama:v$(VERSION)
	docker push registry.deflinhec.dev/nakama:latest

.PHONY: pluginbuilder-image
pluginbuilder-image:
	docker build $(CURDIR) \
		--file build/pluginbuilder/Dockerfile \
		--build-arg "version=$(VERSION)" \
		--build-arg "commit=$(BUILD)" \
		--tag registry.deflinhec.dev/nakama-pluginbuilder:latest

.PHONY: generate
generate:
	go generate -x ./... && \
	(cd console/ui && npm clean-install && npm run-script build)

.PHONY: upgrade
upgrade: export GOPRIVATE=gitlab.com/casino543
upgrade: 
	go get gitlab.com/casino543/nakama-web && \
	go get gitlab.com/casino543/nakama-api && \
	go mod tidy && go mod vendor

default: image plugin-builder-image




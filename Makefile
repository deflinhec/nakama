# Define
VERSION=0.2.1
BUILD=$(shell git rev-parse HEAD)

.PHONY: image
image:
	docker build $(CURDIR) \
		--file build/Dockerfile.local \
		--build-arg "version=$(VERSION)" \
		--build-arg "commit=$(BUILD)" \
		--tag registry.deflinhec.dev/nakama:latest

publish:
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
	(cd web/ui && npm clean-install && npm run-script build) && \
	(cd console/ui && npm clean-install && npm run-script build)

default: image plugin-builder-image




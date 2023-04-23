# Define
VERSION=0.2.6
BUILD=$(shell git rev-parse HEAD)

.PHONY: image
image:
	docker build $(CURDIR) \
		--file build/Dockerfile.local \
		--build-arg "version=$(VERSION)" \
		--build-arg "commit=$(BUILD)" \
		--tag registry.deflinhec.dev/nakama:v$(VERSION)

publish:
	docker tag registry.deflinhec.dev/nakama:v$(VERSION) registry.deflinhec.dev/nakama:latest
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
upgrade:
	GOPRIVATE=gitlab.com/casino543 \
	go get gitlab.com/casino543/nakama-web && \
	go mod tidy && go mod vendor

default: image plugin-builder-image




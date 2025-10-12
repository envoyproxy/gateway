tools.dir = tools
tools.bindir = tools/bin
tools.srcdir = tools/src

$(tools.bindir)/%: $(tools.srcdir)/%/pin.go $(tools.srcdir)/%/go.mod
	cd $(<D) && GOOS= GOARCH= go build -o $(abspath $@) $$(sed -En 's,^import _ "(.*)".*,\1,p' pin.go)

IMAGE_PREFIX ?= nareddyt/
APP_NAME ?= kubecon-2025-fake-spire-server
TAG ?= 2.0.0

PLATFORMS ?= linux/amd64,linux/arm64

.PHONY: docker-buildx
docker-buildx:
	docker buildx build --platform $(PLATFORMS) -f tools/docker/fake-spire-agent/Dockerfile . -t $(IMAGE_PREFIX)$(APP_NAME):$(TAG) --build-arg GO_LDFLAGS="$(GO_LDFLAGS)" --push

build:
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/fake-spire-agent ./cmd/fake-spire-agent

image: build
	docker build -t fake-spire-agent:latest -f tools/docker/fake-spire-agent/Dockerfile .

tools.clean: # Remove all tools
	@$(LOG_TARGET)
	rm -rf $(tools.bindir)

clean: tools.clean
	rm -fr bin


.PHONY: build image tools.clean clean

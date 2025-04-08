tools.dir = tools
tools.bindir = tools/bin
tools.srcdir = tools/src

$(tools.bindir)/%: $(tools.srcdir)/%/pin.go $(tools.srcdir)/%/go.mod
	cd $(<D) && GOOS= GOARCH= go build -o $(abspath $@) $$(sed -En 's,^import _ "(.*)".*,\1,p' pin.go)

IMAGE_PREFIX ?= envoyproxy/gateway-
APP_NAME ?= extension-server
TAG ?= latest

.PHONY: docker-buildx
docker-buildx: build
	docker buildx build -f tools/docker/extension-server/Dockerfile . -t $(IMAGE_PREFIX)$(APP_NAME):$(TAG) --build-arg GO_LDFLAGS="$(GO_LDFLAGS)" --load

build: generate manifests
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/extension-server ./cmd/extension-server

image: build
	docker build -t extension-server:latest -f tools/docker/extension-server/Dockerfile .

manifests:
	@go tool controller-gen crd:allowDangerousTypes=true paths="./..." output:crd:artifacts:config=charts/extension-server/crds/generated

generate:
	@go tool controller-gen object:headerFile="$(tools.dir)/boilerplate.generatego.txt",year=2024 paths="{./api/...}"

tools.clean: # Remove all tools
	@$(LOG_TARGET)
	rm -rf $(tools.bindir)

clean: tools.clean
	rm -fr bin


.PHONY: build image manifests generate tools.clean clean

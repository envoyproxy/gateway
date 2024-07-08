tools.bindir = tools/bin
tools.srcdir = tools/src

tools/controller-gen     = $(tools.bindir)/controller-gen
$(tools.bindir)/%: $(tools.srcdir)/%/pin.go $(tools.srcdir)/%/go.mod
	cd $(<D) && GOOS= GOARCH= go build -o $(abspath $@) $$(sed -En 's,^import _ "(.*)".*,\1,p' pin.go)

build: generate manifests
	mkdir -p bin
	CGO_ENABLED=0 go build -o bin/extension-server ./cmd/extension-server

image: build
	docker build -t extension-server:latest -f tools/docker/extension-server/Dockerfile .

manifests: $(tools/controller-gen)
	$(tools/controller-gen) crd:allowDangerousTypes=true paths="./..." output:crd:artifacts:config=charts/extension-server/crds/generated

generate: $(tools/controller-gen)
	$(tools/controller-gen) object paths="{./api/...}"

tools.clean: # Remove all tools
	@$(LOG_TARGET)
	rm -rf $(tools.bindir)

clean: tools.clean
	rm -fr bin


.PHONY: build image manifests generate tools.clean clean

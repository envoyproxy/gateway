tools.bindir = tools/bin
tools.srcdir = tools/src

# Shell scripts
# =============
#
tools/whitenoise = $(tools.bindir)/whitenoise
$(tools.bindir)/%: $(tools.srcdir)/%.sh
	mkdir -p $(@D)
	install $< $@

# `go get`-able things
# ====================
#
tools/applyconfiguration-gen = $(tools.bindir)/applyconfiguration-gen
tools/buf                    = $(tools.bindir)/buf
tools/client-gen             = $(tools.bindir)/client-gen
tools/controller-gen         = $(tools.bindir)/controller-gen
tools/crd-ref-docs           = $(tools.bindir)/crd-ref-docs
tools/gci                    = $(tools.bindir)/gci
tools/golangci-lint          = $(tools.bindir)/golangci-lint
tools/helm-docs              = $(tools.bindir)/helm-docs
tools/informer-gen           = $(tools.bindir)/informer-gen
tools/jb                     = $(tools.bindir)/jb
tools/jsonnet                = $(tools.bindir)/jsonnet
tools/kind                   = $(tools.bindir)/kind
tools/kustomize              = $(tools.bindir)/kustomize
tools/lister-gen             = $(tools.bindir)/lister-gen
tools/openapi-gen            = $(tools.bindir)/openapi-gen
tools/protoc-gen-go          = $(tools.bindir)/protoc-gen-go
tools/protoc-gen-go-grpc     = $(tools.bindir)/protoc-gen-go-grpc
tools/register-gen           = $(tools.bindir)/register-gen
# Override register-gen target with custom behavior
tools/setup-envtest          = $(tools.bindir)/setup-envtest
$(tools.bindir)/%: $(tools.srcdir)/%/pin.go $(tools.srcdir)/%/go.mod
	cd $(<D) && GOOS= GOARCH= go build -o $(abspath $@) $$(sed -En 's,^import _ "(.*)".*,\1,p' pin.go)

$(tools.bindir)/register-gen: $(tools.srcdir)/register-gen/pin.go $(tools.srcdir)/register-gen/go.mod $(tools.srcdir)/register-gen/fix.patch
	cd $(<D) && \
		go mod vendor && \
		patch -p1 < fix.patch && \
		GOOS= GOARCH= go build -o $(abspath $@) $$(sed -En 's,^import _ "(.*)".*,\1,p' pin.go) && \
		rm -rf vendor


# `pip install`-able things
# =========================
#
tools/codespell    = $(tools.bindir)/codespell
tools/yamllint     = $(tools.bindir)/yamllint
tools/sphinx-build = $(tools.bindir)/sphinx-build
tools/release-notes-docs = $(tools.bindir)/release-notes-docs
$(tools.bindir)/%.d/venv: $(tools.srcdir)/%/requirements.txt
	mkdir -p $(@D)
	python3 -m venv $@
	$@/bin/pip3 install -r $< || (rm -rf $@; exit 1)
$(tools.bindir)/%: $(tools.bindir)/%.d/venv	
	@if [ -e $(tools.srcdir)/$*/$*.sh ]; then \
		ln -sf ../../$(tools.srcdir)/$*/$*.sh $@; \
	else \
		ln -sf $*.d/venv/bin/$* $@; \
	fi

ifneq ($(GOOS),windows)
# Shellcheck
# ==========
#
tools/shellcheck = $(tools.bindir)/shellcheck
SHELLCHECK_VERSION=0.8.0
SHELLCHECK_ARCH=$(shell uname -m)
# shellcheck uses the same binary on Intel and Apple Silicon Mac.
ifeq ($(GOOS),darwin)
SHELLCHECK_ARCH=x86_64
endif
SHELLCHECK_TXZ = https://github.com/koalaman/shellcheck/releases/download/v$(SHELLCHECK_VERSION)/shellcheck-v$(SHELLCHECK_VERSION).$(GOOS).$(SHELLCHECK_ARCH).tar.xz
tools/bin/$(notdir $(SHELLCHECK_TXZ)):
	mkdir -p $(@D)
	curl -sfL $(SHELLCHECK_TXZ) -o $@
%/bin/shellcheck: %/bin/$(notdir $(SHELLCHECK_TXZ))
	mkdir -p $(@D)
	tar -C $(@D) -Jxmf $< --strip-components=1 shellcheck-v$(SHELLCHECK_VERSION)/shellcheck
endif

tools.clean: # Remove all tools
	@$(LOG_TARGET)
	rm -rf $(tools.bindir)

.PHONY: clean
clean: ## Remove all files that are created during builds.
clean: tools.clean

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
define go_install
    GOOS= GOARCH= GOBIN=$(ROOT_DIR)/$(tools.bindir) go install $(1)
endef

tools/controller-gen     = $(tools.bindir)/controller-gen
tools/golangci-lint      = $(tools.bindir)/golangci-lint
tools/kustomize          = $(tools.bindir)/kustomize
tools/kind               = $(tools.bindir)/kind
tools/setup-envtest      = $(tools.bindir)/setup-envtest
tools/crd-ref-docs = $(tools.bindir)/crd-ref-docs
tools/buf                = $(tools.bindir)/buf
tools/protoc-gen-go      = $(tools.bindir)/protoc-gen-go
tools/protoc-gen-go-grpc = $(tools.bindir)/protoc-gen-go-grpc
tools/helm-docs          = $(tools.bindir)/helm-docs

$(tools.bindir)/controller-gen:
	$(call go_install,sigs.k8s.io/controller-tools/cmd/controller-gen@v0.13.0)

$(tools.bindir)/golangci-lint:
	$(call go_install,github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2)

$(tools.bindir)/kind:
	$(call go_install,sigs.k8s.io/kind@v0.20.0)

$(tools.bindir)/setup-envtest:
	$(call go_install,sigs.k8s.io/controller-runtime/tools/setup-envtest@v0.0.0-20220706173534-cd0058ad295c)

$(tools.bindir)/crd-ref-docs:
	$(call go_install,github.com/elastic/crd-ref-docs@v0.0.9)

$(tools.bindir)/buf:
	$(call go_install,github.com/bufbuild/buf/cmd/buf@v1.28.1)

$(tools.bindir)/protoc-gen-go:
	$(call go_install,google.golang.org/protobuf/cmd/protoc-gen-go@v1.30.0)

$(tools.bindir)/protoc-gen-go-grpc:
	$(call go_install,google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0)

$(tools.bindir)/helm-docs:
	$(call go_install,github.com/norwoodj/helm-docs/cmd/helm-docs@v1.12.0)

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

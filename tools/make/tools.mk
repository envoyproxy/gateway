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
tools/controller-gen = $(tools.bindir)/controller-gen
tools/golangci-lint  = $(tools.bindir)/golangci-lint
tools/kustomize      = $(tools.bindir)/kustomize
tools/setup-envtest  = $(tools.bindir)/setup-envtest
$(tools.bindir)/%: $(tools.srcdir)/%/pin.go $(tools.srcdir)/%/go.mod
	cd $(<D) && GOOS= GOARCH= go build -o $(abspath $@) $$(sed -En 's,^import "(.*)".*,\1,p' pin.go)

# `pip install`-able things
# =========================
#
tools/codespell = $(tools.bindir)/codespell
tools/yamllint  = $(tools.bindir)/yamllint
$(tools.bindir)/%.d/venv: $(tools.srcdir)/%/requirements.txt
	mkdir -p $(@D)
	python3 -m venv $@
	$@/bin/pip3 install -r $< || (rm -rf $@; exit 1)
$(tools.bindir)/%: $(tools.bindir)/%.d/venv
	ln -sf $*.d/venv/bin/$* $@

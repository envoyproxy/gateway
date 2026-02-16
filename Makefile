# All make targets should be implemented in tools/make/*.mk
# ====================================================================================================
# Supported Targets: (Run `make help` to see more information)
# ====================================================================================================

# This file is a wrapper around `make` so that we can force on the
# --warn-undefined-variables flag.  Sure, you can set
# `MAKEFLAGS += --warn-undefined-variables` from inside of a Makefile,
# but then it won't turn on until the second phase (recipe execution),
# and won't actually be on during the initial phase (parsing).
# See: https://www.gnu.org/software/make/manual/make.html#Reading-Makefiles

# Have everything-else ("%") depend on _run (which uses
# $(MAKECMDGOALS) to decide what to run), rather than having
# everything else run $(MAKE) directly, since that'd end up running
# multiple sub-Makes if you give multiple targets on the CLI.
_run:
	@$(MAKE) --warn-undefined-variables -f tools/make/common.mk $(MAKECMDGOALS)
.PHONY: _run
$(if $(MAKECMDGOALS),$(MAKECMDGOALS): %: _run)

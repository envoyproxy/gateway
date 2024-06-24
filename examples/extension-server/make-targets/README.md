# Make Build System

This directory contains all of the make targets for the build system

Each file is tied to a category under `make help` via the `##@ Category` line at the top of the file

If you want to add more make files, just add a tag like that at the top of the file so it gets added to the output from
`make help`

- Public consumption targets should contain a docstring next to the target starting with `##`
  so that `make help` can automatically display the target and what it does.

  Example:

  ```makefile
  .PHONY foo
  foo: ## does magical things
  ```

- Intermediate targets not meant for public consumption should begin with an underscore `_`

  Example:

  ```makefile
  .PHONY _dependency_target
  _dependency_target:
    echo "do some cool stuff"
  ```

- mk files that do not contain targets for public consumption should also begin with an underscore `_`.

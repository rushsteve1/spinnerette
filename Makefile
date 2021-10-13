##
# Spinnerette
#
# @file
# @version 0.1

# Named rule aliases
all: spinnerette
submodule: deps/janet/README.md
janet: deps/janet/build/c/janet.c

spinnerette: deps/janet/build/c/janet.c
	go build

deps/janet/build/c/janet.c: deps/janet/README.md
	$(MAKE) -C deps/janet

# This rule just makes sure the README exists
# which means the submodule has been pulled
deps/janet/README.md:
	git submodule update --init --recursive

.PHONY: clean clean-janet

clean: clean-janet
	rm -f spinnerette

clean-janet:
	$(MAKE) -C deps/janet clean

# end

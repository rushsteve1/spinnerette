##
# Spinnerette
#
# @file
# @version 0.1

# Named rule aliases
all: spinnerette
submodules: bindings/janet/README.md
janet: bindings/janet/build/c/janet.c
sqlite: bindings/libsqlite3.a


spinnerette: $(wildcard **/*.go) bindings/shim.c bindings/janet/build/libjanet.a bindings/libsqlite3.a
	go build

bindings/sqlite3.o: bindings/sqlite3/sqlite3.c
	$(CC) -fPIC -O2 -c $^ -o $@

bindings/libsqlite3.a: bindings/sqlite3.o
	ar rcs $@ $^

bindings/janet/build/libjanet.a: bindings/janet/README.md
	$(MAKE) -C bindings/janet

# This rule just makes sure the README exists
# which means the submodule has been pulled
bindings/janet/README.md:
	git submodule update --init --recursive


.PHONY: clean clean-janet run

clean: clean-janet
	rm -f spinnerette bindings/sqlite3.o bindings/libsqlite3.a

clean-janet:
	$(MAKE) -C bindings/janet clean

run: spinnerette
	./$<

# end

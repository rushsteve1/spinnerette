##
# Spinnerette
#
# @file
# @version 0.1

# TODO get Janet to use release flags
CFLAGS=-std=c99 -Wall -Wextra -fPIC -O2

GOFILES=$(wildcard **/*.go)
CFILES=$(wildcard bindings/*.c)
JANETFILES=main.go $(wildcard libs/spin/*.janet)
NATIVEFILES=bindings/janet/build/libjanet.a bindings/libsqlite3.a
ALLFILES=$(GOFILES) $(JANETFILES) $(CFILES) $(NATIVEFILES)

# Named rule aliases
all: spinnerette
submodules: bindings/janet/README.md
janet: bindings/janet/build/c/janet.c
sqlite: bindings/libsqlite3.a


spinnerette: $(ALLFILES)
	go build -v ./...
	go build -v

test: $(ALLFILES)
	go test -v ./...

bindings/sqlite3.o: bindings/sqlite3/sqlite3.c
	$(CC) $(CFLAGS) -DSQLITE_ENABLE_JSON1 -c $^ -o $@

bindings/libsqlite3.a: bindings/sqlite3.o
	$(AR) rcs $@ $^

bindings/janet/build/libjanet.a: bindings/janet/README.md
	$(MAKE) -C bindings/janet

# This rule just makes sure the README exists
# which means the submodule has been pulled
bindings/janet/README.md:
	git submodule update --init --recursive


.PHONY: clean clean-janet run

clean: clean-janet clean-sqlite
	$(RM) -f spinnerette

clean-janet:
	$(MAKE) -C bindings/janet clean

clean-sqlite:
	$(RM) -f bindings/sqlite3.o bindings/libsqlite3.a

run: spinnerette
	./$<

# end

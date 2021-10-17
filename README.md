# Spinnerette

Simple Janet web development platform in Go, à la PHP

## Building

```sh
make
```

This will handle pulling submodules, building Janet, and then building and
linking Spinnerette

If you are hacking on Spinnerette, once `make` has been called once switch to
using `go build`.

## Usage

With the binary built, run the following command:

```sh
./spinnerette
```

Which will start a server on port 9999 then you may visit the examples like the
following:

[http://localhost:9999/examples/hello.janet](http://localhost:9999/examples/hello.janet)

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

The [Makefile](./Makefile) does provide a shortcut function for development. Run
the following command to compile the `spinnerette` binary and spin up the web
server:

```sh
make run
```

## CLI Args

The `spinnerette` binary accepts the following arguments:

- `--method string`
The method that Spinnerette will listen on (HTTP, FastCGI, or CGI) (default "http")

- `--port int`
Port to use for HTTP/FastCGI (default 9999)

- `--root string`
Webroot files will be found in (default your current working directory)

- `--socket string`
Socket to use for FastCGI (falls back to TCP with --port)

### Example

```sh
./spinnerette -port 3000 -root ./examples/
```

## How it Works

The Spinnerette binary starts a webserver that can execute
[janet](https://janet-lang.org) files as web pages, similar to PHP but with the
sweet, sweet goodness of a modern lisp language inspired by languages like
Clojure.

The goal is to allow spinnerette to run on the cheapest of shared web hosts to
support rapidly building server-side scripts.

Spinnerette can works with just about any frontend by adding the necessary
script tags in the response of your janet or temple files.

See the [examples](./examples) for a look at how it all works.

## Where does the name Spinnerette come from?

Spinnerette is a play on the silk-spinning organ spiders possess to create their
intricate webs fast.

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

The [Makefile](./Makefile) does provide a shortcut function for development. Run the following
command to compile the `spinnerette` binary and spin up the web server:

```sh
make run
```

## CLI Args

The spinnerette binary accepts the following arguments:

### -method string

The method that Spinnerette will listen on (HTTP, FastCGI, or CGI) (default "http")

### -port int

Port to use for HTTP/FastCGI (default 9999)

### -root string

Webroot files will be found in (default "/Users/j/projects/spinnerette")

### -socket string

Socket to use for FastCGI (falls back to TCP with --port)

### Example

```sh
./spinnerette -port 3000 -root ./examples/
```

## How it Works

The Spinnerette binary starts a webserver that can execute [janet](https://janet-lang.org) files as web pages, similar
to PHP but with the sweet, sweet goodness of a modern lisp language inspired by
languages like Clojure.

The goal is to allow spinnerette to run on the cheapest of shared web hosts to
support rapidly building server-side scripts.

Spinnerette can works with just about any frontend by adding the necessary
script tags in the response of your janet or temple files.

### Janet Strings

For very simple use cases, janet files may return a string which will be
directly output to incoming browser requests.

Try visiting [http://localhost:9999/examples/hello.janet] to see it in action.

```janet
# The last value returned by the script will be used as the response
# This value can be either a string, or a Circlet-style response object

# Simply responds with the string
"Hello World!"
```

### Janet HTML

The janet/html library is already included which supports hiccup like syntax.
Try visiting [http://localhost:9999/examples/html.janet] which has code like the following:

```janet
# Spinnerette bundles in the janet-html library for easily creating HTML pages
# with pure Janet. It uses a syntax similar to Clojure's Hiccup

(import html)

(html/encode
 [:html
  [:body
   [:h1 "Hello from Janet-HTML"]
   [:p "this was created with pure Janet!"]]])

```

The `html/encode` function takes the hiccup-like data structure and transforms
it into an html string which is then returned as the browser response.

### Temple

Temple files work more similarly to a PHP script support interweaving Janet code
within traditional HTML.

Try visiting [http://localhost:9999/examples/hello.temple] which contains code
like the following:

```temple
<!-- Temple is also supported via the Spork library -->
<!-- These templates can have Janet code mixed into them in a few ways-->

<!-- This will be evaluated at compile-time -->
{$ (import html) $}

<html>
    <body>
      <!-- This will be evaluated and its return value escaped and put inline -->
      <h1>{{ "Hello there!" }}</h1>

      <!-- This will be evaluated and NOT escaped -->
      <!-- You can mix janet-html into Temple! -->
      {- (html/encode [:div [:p "Fun with templates"]]) -}

      <!-- This will be evaluated but NOT added to the page -->
      {% "I don't go anywhere" %}

      <!-- Printing in Temple also adds to the page -->
      {% (print "I get added anyway!") %}
    </body>
</html>
```

`{- janetcode}` evaluates janet code in-place and interpolates the return value.
In this case `html/encode` is transforming a hiccup-like tree of html tags into
a single HTML string and rendering it in place within the body tag.

## Where does the name Spinnerette come from?

Spinnerette is a play on the silk-spinning organ spiders possess to create their
intricate webs fast.

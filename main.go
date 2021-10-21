package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"mime"
	"net"
	"net/http"
	"net/http/cgi"
	"net/http/fcgi"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	janet "github.com/rushsteve1/spinnerette/bindings"
)

type Flags struct {
	Root   string
	Method string
	Port   int
	Socket string
}

var parsedFlags Flags

//go:embed libs/janet-html/src/janet-html.janet libs/spork/spork/*.janet libs/spin/*.janet
var embeddedLibs embed.FS

func main() {
	runtime.GOMAXPROCS(1)

	ParseFlags()
	parsedFlags.Method = strings.ToLower(parsedFlags.Method)

	// Setup the Janet interpreter
	janet.SetupEmbeds(embeddedLibs)
	janet.StartJanet()
	defer janet.StopJanet()

	// Add mimetypes to database
	mime.AddExtensionType(".janet", "text/janet")
	mime.AddExtensionType(".temple", "text/temple")

	handler := Handler{
		Addr: fmt.Sprintf("0.0.0.0:%d", parsedFlags.Port),
	}

	if parsedFlags.Method == "http" {
		log.Printf("Starting HTTP server at http://%s", handler.Addr)
		http.ListenAndServe(handler.Addr, handler)
	} else if parsedFlags.Method == "fastcgi" || parsedFlags.Method == "fcgi" {
		var listen net.Listener
		defer listen.Close()

		var err error
		if len(parsedFlags.Socket) > 0 {
			log.Printf("Starting FastCGI server on socket %s", parsedFlags.Socket)
			listen, err = net.Listen("unix", parsedFlags.Socket)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Printf("Starting FastCGI server on port %d", parsedFlags.Port)
			listen, err = net.Listen("tcp", handler.Addr)
			if err != nil {
				log.Fatal(err)
			}
		}

		fcgi.Serve(listen, handler)
	} else if parsedFlags.Method == "cgi" {
		log.Printf("Running as CGI program")
		cgi.Serve(handler)
	} else {
		log.Fatal("Unknown method")
	}
}

func ParseFlags() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&parsedFlags.Method, "method", "http", "The method that Spinnerette will listen on (HTTP, FastCGI, or CGI)")
	flag.StringVar(&parsedFlags.Root, "root", wd, "Webroot files will be found in")
	flag.IntVar(&parsedFlags.Port, "port", 9999, "Port to use for HTTP/FastCGI")
	flag.StringVar(&parsedFlags.Socket, "socket", "", "Socket to use for FastCGI (falls back to TCP with --port)")

	flag.Parse()
}

type Handler struct {
	Addr string
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(parsedFlags.Root, filepath.Clean(r.URL.Path))

	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	m := mime.TypeByExtension(filepath.Ext(path))
	switch m {
	case "text/janet; charset=utf-8":
		h.janetHandler(w, r, path)
	case "text/temple; charset=utf-8":
		h.templeHandler(w, r, path)
	default:
		http.ServeFile(w, r, path)
	}
}

func (h Handler) janetHandler(w http.ResponseWriter, r *http.Request, path string) {
	j, err := janet.EvalFilePath(path, r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Println(err.Error())
		return
	}

	if j != nil {
		janet.WriteResponse(*j, w)
	}
}

func (h Handler) templeHandler(w http.ResponseWriter, r *http.Request, path string) {
	j, err := janet.RenderTemple(path, r)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Println(err.Error())
		return
	}

	if j != nil {
		janet.WriteResponse(*j, w)
	}
}

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/cgi"
	"net/http/fcgi"
	"os"
	"path/filepath"
	"strings"
)

type Flags struct {
	Root   string
	Method string
	Port   int
	Socket string
}

var parsedFlags Flags

func main() {
	parsedFlags = ParseFlags()
	parsedFlags.Method = strings.ToLower(parsedFlags.Method)

	handler := Handler{
		Addr: fmt.Sprintf("0.0.0.0:%d", parsedFlags.Port),
	}

	if parsedFlags.Method == "http" {
		log.Printf("Starting HTTP server on port %d", parsedFlags.Port)
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

func ParseFlags() Flags {
	var flg Flags

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&flg.Method, "method", "http", "The method that Spinnerette will listen on (HTTP, FastCGI, or CGI)")
	flag.StringVar(&flg.Root, "root", wd, "Webroot files will be found in")
	flag.IntVar(&flg.Port, "port", 9999, "Port to use for HTTP/FastCGI")
	flag.StringVar(&flg.Socket, "socket", "", "Socket to use for FastCGI (falls back to TCP with --port)")

	flag.Parse()

	return flg
}

type Handler struct {
	Addr string
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(parsedFlags.Root, r.URL.Path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	switch filepath.Ext(path) {
	case ".janet":
		h.janetHandler(w, r, path)
	case ".temple":
		h.templeHandler(w, r, path)
	default:
		http.ServeFile(w, r, path)
	}
}

func (h Handler) janetHandler(w http.ResponseWriter, r *http.Request, path string) {
	Init()
	defer DeInit()

	env, err := RequestEnv(r)
	if err != nil {
		http.Error(w, "Could not build request env", 500)
		log.Println(err)
		return
	}

	janet, err := EvalFilePath(path, env)
	if err != nil {
		http.Error(w, err.Error(), 500)
		log.Println(err.Error())
		return
	}

	WriteResponse(janet, w)
}

func (h Handler) templeHandler(w http.ResponseWriter, r *http.Request, path string) {
	Init()
	defer DeInit()

	/*
		env, err := RequestEnv(r)
		if err != nil {
			http.Error(w, "Could not build request env", 500)
			log.Println(err)
			return
		}
	*/
}

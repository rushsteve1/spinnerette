package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/cgi"
	"os"
	"path/filepath"
	"strings"
)

type Flags struct {
	Root string
	Method string
	Port int
}

var parsedFlags Flags

func main() {
	parsedFlags = ParseFlags()
	parsedFlags.Method = strings.ToLower(parsedFlags.Method)

	var handler Handler

	if parsedFlags.Method == "http" {
		log.Printf("Starting HTTP server on port %d", parsedFlags.Port)
		addr := fmt.Sprintf(":%d", parsedFlags.Port)
		http.ListenAndServe(addr, handler)
	} else if parsedFlags.Method == "fastcgi" || parsedFlags.Method == "fcgi" {
		// TODO
	} else if parsedFlags.Method == "cgi" {
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

	flag.StringVar(&flg.Method, "method", "http", "The method that Spinnerette will listen on (HTTP, FastCGI, or CGI")
	flag.StringVar(&flg.Root, "root", wd, "Webroot files will be found in")
	flag.IntVar(&flg.Port, "port", 9999, "Port to use for HTTP")

	return flg
}

type Handler struct {}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join(parsedFlags.Root, r.URL.Path)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.NotFound(w, r)
	} else if filepath.Ext(path) == ".janet" {
		janet, err := EvalFilePath(path)
		defer DeInit()
		if err != nil {
			log.Printf(err.Error())
		}
		w.Write([]byte(ToString(janet)))
	} else {
		http.ServeFile(w, r, path)
	}
}
package main

// #include "janet.h"
import "C"
import (
	"embed"
	"fmt"
)

/*
 * This implements embedded module loading via an injected loader
 * an alternative approach might be to use the modules cache
 * using the cache has a higher up-front cost since everything must be loaded
 */

//go:embed deps/janet-html/src/janet-html.janet deps/spork/spork/*.janet
var libs embed.FS

var fileMappings = map[string]string{
	"html":         "deps/janet-html/src/janet-html.janet",
	"spork/http":   "deps/spork/spork/http.janet",
	"spork/temple": "deps/spork/spork/temple.janet",
}

//export ModuleLoader
func ModuleLoader(path *C.char) *C.JanetTable {
	name := C.GoString(path)
	env := CoreEnv()

	if val, ok := fileMappings[name]; ok {
		code, _ := libs.ReadFile(val)
		EvalBytes(code, env)
	}

	return env
}

func InitModules(env *C.JanetTable) {
	keys := ""
	for k := range fileMappings {
		keys += fmt.Sprintf("\"%s\" ", k)
	}

	// This does a little bit of meta-programming trickery to create the predicate
	// function that validates built-in Spinnerette library paths
	pred := fmt.Sprintf("(fn [path] (find |(= path $) [%s] nil))", keys)
	identity, _ := EvalString(pred, env)
	tuple := []C.Janet{identity, jkey("spinnerette")}

	paths := getEnvValue(env, "module/paths")
	C.janet_array_push(C.janet_unwrap_array(paths), C.janet_wrap_tuple(C.janet_tuple_n(&tuple[0], 2)))

	loaders := getEnvValue(env, "module/loaders")
	C.janet_checktype(paths, C.JANET_TABLE)
	fn := getEnvValue(env, "spin/module-loader")
	C.janet_table_put(C.janet_unwrap_table(loaders), jkey("spinnerette"), fn)
}

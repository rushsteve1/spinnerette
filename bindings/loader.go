package bindings

// #include "janet.h"
// #include "shared.h"
import "C"
import (
	"embed"
	"strings"
	"unsafe"
)

/*
 * This implements embedded module loading via an injected loader
 * an alternative approach might be to use the modules cache
 * using the cache has a higher up-front cost since everything must be loaded
 */

var embedded embed.FS

var fileMappings = map[string]string{
	"html":             "libs/janet-html/src/janet-html.janet",
	"spin":             "libs/spin/init.janet",
	"spin/cache":       "libs/spin/cache.janet",
	"spin/response":    "libs/spin/response.janet",
	"spork":            "libs/spork/spork/init.janet",
	"spork/argparse":   "libs/spork/spork/argparse.janet",
	"spork/ev-utils":   "libs/spork/spork/ev-utils.janet",
	"spork/fmt":        "libs/spork/spork/fmt.janet",
	"spork/generators": "libs/spork/spork/generators.janet",
	"spork/http":       "libs/spork/spork/http.janet",
	"spork/init":       "libs/spork/spork/init.janet",
	"spork/misc":       "libs/spork/spork/misc.janet",
	"spork/msg":        "libs/spork/spork/msg.janet",
	"spork/netrepl":    "libs/spork/spork/netrepl.janet",
	"spork/path":       "libs/spork/spork/path.janet",
	"spork/regex":      "libs/spork/spork/regex.janet",
	"spork/rpc":        "libs/spork/spork/rpc.janet",
	"spork/temple":     "libs/spork/spork/temple.janet",
	"spork/test":       "libs/spork/spork/test.janet",
}

var nativeModules = []string{"json", "sqlite3"}
var prefixes = []string{"spin", "spork"}

// This is a workaround for how Go'd embed works
func SetupEmbeds(e embed.FS) {
	embedded = e
}

//export moduleLoader
func moduleLoader(p *C.char, protoEnv *C.JanetTable) *C.JanetTable {
	path := C.GoString(p)
	env := C.janet_table(1024)
	env.proto = protoEnv

	switch path {
	case "json":
		C.janet_cfuns(env, C.CString("json"), (*C.JanetReg)(unsafe.Pointer(C.json_ns.cfuns)))
	case "sqlite3":
		C.janet_cfuns(env, C.CString("sqlite3"), (*C.JanetReg)(unsafe.Pointer(C.sqlite3_ns.cfuns)))
	default:
		if val, ok := fileMappings[path]; ok {
			code, _ := embedded.ReadFile(val)
			EvalBytes(code, val, nil)
		}
	}

	return env
}

// TODO handle relative paths in bundled modules

//export pathPred
func pathPred(j C.Janet) C.Janet {
	path := C.GoString((*C.char)(unsafe.Pointer(C.janet_unwrap_string(j))))

	for _, s := range nativeModules {
		if s == path {
			return j
		}
	}
	for s := range fileMappings {
		if s == path {
			return j
		}
	}

	// TODO this currently allows for user scripts to access bundled libraries
	// with relative paths when it probably shouldn't

	// Tries to load relative imports to fix init.janet files
	if strings.HasPrefix(path, "./") {
		for _, prefix := range prefixes {
			s := strings.Replace(path, ".", prefix, 1)
			if _, ok := fileMappings[s]; ok {
				return jstr(s)
			}
		}
	}

	return jnil()
}

func initModules() *C.JanetTable {
	env := C.janet_core_env((*C.JanetTable)(C.NULL))
	// Load the shim functions
	// When loading them link this there will be no prefix
	// So the prefix is adde in spin_cfuns
	C.janet_cfuns(env, C.CString(""), (*C.JanetReg)(unsafe.Pointer(&C.spin_cfuns)))

	bindToEnv(env, "*cache*", C.janet_wrap_table(C.janet_table(0)), "Internal cache table. Use `spin/cache` instead.")

	pred, _ := EvalString("(fn [path] (spinternal/path-pred path))")
	tuple := []C.Janet{pred, jkey("spinnerette")}

	paths := getEnvValue(env, "module/paths")
	C.janet_array_push(C.janet_unwrap_array(paths), C.janet_wrap_tuple(C.janet_tuple_n(&tuple[0], 2)))

	loaders := getEnvValue(env, "module/loaders")
	C.janet_checktype(paths, C.JANET_TABLE)
	fn := C.janet_wrap_cfunction(C.JanetCFunction(C.loader_shim))
	C.janet_table_put(C.janet_unwrap_table(loaders), jkey("spinnerette"), fn)

	return env
}

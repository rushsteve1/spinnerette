package bindings

// #include "janet.h"
// #include "shared.h"
import "C"
import (
	"embed"
	"log"
	"unsafe"
)

/*
 * This implements embedded module loading via an injected loader
 * an alternative approach might be to use the modules cache
 * using the cache has a higher up-front cost since everything must be loaded
 */

var embedded embed.FS

var fileMappings = map[string]string{
	"html": "libs/janet-html/src/janet-html.janet",

	"spin/cache":    "libs/spin/cache.janet",
	"spin/response": "libs/spin/response.janet",
	"spin":          "libs/spin/init.janet",

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
	"spork":            "libs/spork/spork/init.janet",
}

var nativeModules = []string{"json", "sqlite3"}
var prefixes = []string{"spin", "spork"}

// This is a workaround for how Go'd embed works
func SetupEmbeds(e embed.FS) {
	embedded = e
}

func moduleLoader(path string, protoEnv *C.JanetTable) C.Janet {
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
			_, err := innerEval(env, code, val)
			if err != nil {
				log.Println(err.Error())
			}
		}
	}

	return C.janet_wrap_table(env)
}

// Sets up the environment and eagerly loads all bundled modules into
// module/cache
func initModules() *C.JanetTable {
	env := C.janet_core_env((*C.JanetTable)(C.NULL))

	// Load the spinternal module into the global environment
	C.janet_cfuns_ext_prefix(env, C.CString("spinternal"), (*C.JanetRegExt)(unsafe.Pointer(&C.spin_cfuns)))

	// The internal cache used by spin/cache
	bindToEnv(env, "spinternal/cache",
		C.janet_wrap_table(C.janet_table(0)),
		"Internal cache table. Use `spin/cache` to access.",
	)

	moduleCache := C.janet_unwrap_table(envResolve(env, "module/cache"))

	// TODO relative imports within imports

	for _, s := range nativeModules {
		C.janet_table_put(moduleCache, jstr(s), moduleLoader(s, env))
	}
	for s := range fileMappings {
		C.janet_table_put(moduleCache, jstr(s), moduleLoader(s, env))
	}

	return env
}

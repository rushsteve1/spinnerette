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

var filePaths = []string{
	"libs/janet-html/src/janet-html.janet",

	"libs/spin/cache.janet",
	"libs/spin/responses.janet",
	"libs/spin/init.janet",

	"libs/spork/spork/msg.janet",
	"libs/spork/spork/argparse.janet",
	"libs/spork/spork/ev-utils.janet",
	"libs/spork/spork/fmt.janet",
	"libs/spork/spork/generators.janet",
	"libs/spork/spork/http.janet",
	"libs/spork/spork/misc.janet",
	"libs/spork/spork/netrepl.janet",
	"libs/spork/spork/path.janet",
	"libs/spork/spork/regex.janet",
	"libs/spork/spork/rpc.janet",
	"libs/spork/spork/temple.janet",
	"libs/spork/spork/test.janet",
	"libs/spork/spork/init.janet",
}

const startupPath string = "libs/startup.janet"

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
		code, _ := embedded.ReadFile(path)
		_, err := innerEval(env, code, path)
		if err != nil {
			log.Println(err.Error())
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

	code, err := embedded.ReadFile(startupPath)
	if err != nil {
		log.Fatal(err)
	}
	innerEval(env, code, startupPath)
	log.Println("Evaluated startup.janet")

	moduleCache := C.janet_unwrap_table(envResolve(env, "module/cache"))

	// TODO relative imports within imports

	for _, s := range nativeModules {
		C.janet_table_put(moduleCache, jstr(s), moduleLoader(s, env))
	}
	for _, p := range filePaths {
		C.janet_table_put(moduleCache, jstr(p), moduleLoader(p, env))
	}

	return env
}

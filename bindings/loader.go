package bindings

// #include "janet.h"
// #include "shared.h"
import "C"
import (
	"embed"
	"unsafe"
)

/*
 * This implements embedded module loading via an injected loader
 * an alternative approach might be to use the modules cache
 * using the cache has a higher up-front cost since everything must be loaded
 */

var embedded embed.FS
var fileMappings map[string]string
var nativeModules = []string{"json", "sqlite3"}

// This is a workaround for how Go'd embed works
func SetupEmbeds(e embed.FS, m map[string]string) {
	embedded = e
	fileMappings = m
}

//export ModuleLoader
func ModuleLoader(path *C.char, protoEnv *C.JanetTable) *C.JanetTable {
	name := C.GoString(path)
	env := C.janet_table(1024)
	env.proto = protoEnv

	switch name {
	case "json":
		C.janet_cfuns(env, path, (*C.JanetReg)(unsafe.Pointer(C.json_ns.cfuns)))
	case "sqlite3":
		C.janet_cfuns(env, path, (*C.JanetReg)(unsafe.Pointer(C.sqlite3_ns.cfuns)))
	default:
		if val, ok := fileMappings[name]; ok {
			code, _ := embedded.ReadFile(val)
			EvalBytes(code, env)
		}
	}

	return env
}

// TODO handle relative paths in bundled modules

//export PathPred
func PathPred(path C.Janet) C.Janet {
	p := C.GoString((*C.char)(unsafe.Pointer(C.janet_unwrap_string(path))))
	for _, s := range nativeModules {
		if s == p {
			return path
		}
	}
	for s, _ := range fileMappings {
		if s == p {
			return path
		}
	}

	return C.janet_wrap_nil()
}

func InitModules(env *C.JanetTable) {
	// Load the shim functions
	C.janet_cfuns(env, C.CString(""), (*C.JanetReg)(unsafe.Pointer(&C.shim_cfuns)))

	pred, _ := EvalString("(fn [path] (spin/path-pred path))", env)
	tuple := []C.Janet{pred, jkey("spinnerette")}

	paths := getEnvValue(env, "module/paths")
	C.janet_array_push(C.janet_unwrap_array(paths), C.janet_wrap_tuple(C.janet_tuple_n(&tuple[0], 2)))

	loaders := getEnvValue(env, "module/loaders")
	C.janet_checktype(paths, C.JANET_TABLE)
	fn := C.janet_wrap_cfunction(C.JanetCFunction(C.loader_shim))
	C.janet_table_put(C.janet_unwrap_table(loaders), jkey("spinnerette"), fn)
}

package main

/*
#cgo CFLAGS: -fPIC -O2
#cgo CFLAGS: -I ${SRCDIR}/deps/janet/src/include
#cgo CFLAGS: -I ${SRCDIR}/deps/janet/src/conf
#cgo LDFLAGS: -lm -ldl -lrt -lpthread
#include "deps/janet/build/c/janet.c"
#include "janet.h"
*/
import "C"

func ParentEnv() *C.JanetTable {
	return (*C.JanetTable)(C.NULL)
}

func EvalString(code string) (C.Janet, error) {
	C.janet_init()
	env := C.janet_core_env(ParentEnv())
	var out C.Janet 
	C.janet_dostring(env, C.CString(code), C.CString("spinnerette"), &out)
	C.janet_deinit()
	return out, nil
}

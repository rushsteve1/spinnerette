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
import (
	"errors"
	"fmt"
	"os"
	"unsafe"
)

func DeInit() {
	C.janet_deinit()
}

func ParentEnv() *C.JanetTable {
	return (*C.JanetTable)(C.NULL)
}

func EvalString(code string) (C.Janet, error) {
	C.janet_init()
	env := C.janet_core_env(ParentEnv())
	var out C.Janet 
	errno := C.janet_dostring(env, C.CString(code), C.CString("spinnerette"), &out)
	if errno != 0 {
		return C.janet_wrap_nil(), errors.New(fmt.Sprintf("Janet error: %d", errno))
	}
	return out, nil
}

func EvalBytes(code []byte) (C.Janet, error) {
	C.janet_init()
	env := C.janet_core_env(ParentEnv())
	var out C.Janet
	errno := C.janet_dobytes(env, (*C.uchar)(unsafe.Pointer(&code[0])), (C.int)(len(code)), C.CString("spinnerette"), &out)
	if errno != 0 {
		return C.janet_wrap_nil(), errors.New(fmt.Sprintf("Janet error: %d", errno))
	}
	return out, nil
}

func EvalFilePath(path string) (C.Janet, error) {
	code, err := os.ReadFile(path)
	if err != nil {
		return C.janet_wrap_nil(), err
	}

	C.janet_init()
	env := C.janet_core_env(ParentEnv())
	var out C.Janet
	errno := C.janet_dobytes(env, (*C.uchar)(unsafe.Pointer(&code[0])), (C.int)(len(code)), C.CString(path), &out)
	if errno != 0 {
		return C.janet_wrap_nil(), errors.New(fmt.Sprintf("Janet error: %d", errno))
	}
	return out, nil
}

func ToString(janet C.Janet) string {
	return C.GoString((*C.char)(unsafe.Pointer(C.janet_to_string(janet))))
}
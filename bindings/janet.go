package bindings

/*
#cgo CFLAGS: -fPIC -O2
#cgo CFLAGS: -I ${SRCDIR}/janet/build
#cgo LDFLAGS: -lm -ldl -lpthread ${SRCDIR}/janet/build/libjanet.a ${SRCDIR}/libsqlite3.a

#include "janet.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"os"
	"unsafe"
)

func Init() {
	C.janet_init()
}

func DeInit() {
	C.janet_deinit()
}

func CoreEnv() *C.JanetTable {
	return C.janet_core_env((*C.JanetTable)(C.NULL))
}

func SpinEnv() *C.JanetTable {
	env := CoreEnv()
	InitModules(env)
	return env
}

func EvalFilePath(path string, env *C.JanetTable) (C.Janet, error) {
	code, err := os.ReadFile(path)
	if err != nil {
		return C.janet_wrap_nil(), err
	}

	if len(code) == 0 {
		return C.janet_wrap_nil(), errors.New("File is empty")
	}

	var out C.Janet
	errno := C.janet_dobytes(env, (*C.uchar)(unsafe.Pointer(&code[0])), C.int(len(code)), C.CString(path), &out)
	if errno != 0 {
		return C.janet_wrap_nil(), errors.New(fmt.Sprintf("Janet error: %d", errno))
	}

	return out, nil
}

func EvalBytes(code []byte, env *C.JanetTable) (C.Janet, error) {
	if len(code) == 0 {
		return C.janet_wrap_nil(), errors.New("Code is empty")
	}

	var out C.Janet
	errno := C.janet_dobytes(env, (*C.uchar)(unsafe.Pointer(&code[0])), C.int(len(code)), C.CString("spinnerette internal"), &out)
	if errno != 0 {
		return C.janet_wrap_nil(), errors.New(fmt.Sprintf("Janet error: %d", errno))
	}

	return out, nil
}

func EvalString(code string, env *C.JanetTable) (C.Janet, error) {
	return EvalBytes([]byte(code), env)
}

func EvalBind(env *C.JanetTable, key string, code string, doc string) error {
	j, err := EvalString(code, env)
	if err != nil {
		return err
	}

	bindToEnv(env, key, j, doc)
	return nil
}

func ToString(janet C.Janet) string {
	return C.GoString((*C.char)(unsafe.Pointer(C.janet_to_string(janet))))
}

func PrettyPrint(j C.Janet) string {
	buf := C.janet_pretty((*C.JanetBuffer)(C.NULL), -1, C.JANET_PRETTY_NOTRUNC, j)
	return C.GoStringN((*C.char)(unsafe.Pointer(buf.data)), buf.count)
}

func bindToEnv(env *C.JanetTable, key string, value C.Janet, doc string) {
	C.janet_def_sm(env, C.CString(key), value, C.CString(doc), C.CString("janet.go"), -1)
}

func getEnvValue(env *C.JanetTable, key string) C.Janet {
	table := C.janet_unwrap_table(C.janet_table_get(env, jsym(key)))
	return C.janet_table_get(table, jkey("value"))
}

func jbuf(b []byte) C.Janet {
	buf := C.janet_buffer(C.int(len(b)))
	C.janet_buffer_push_string(buf, (*C.uchar)(unsafe.Pointer(&b[0])))
	return C.janet_wrap_buffer(buf)
}

func jstr(s string) C.Janet {
	cstr := (*C.uchar)(unsafe.Pointer(C.CString(s)))
	str := C.janet_string(cstr, C.int(len(s)))
	return C.janet_wrap_string(str)
}

func jsym(s string) C.Janet {
	cstr := (*C.uchar)(unsafe.Pointer(C.CString(s)))
	sym := C.janet_symbol(cstr, C.int(len(s)))
	return C.janet_wrap_symbol(sym)
}

func jkey(s string) C.Janet {
	cstr := (*C.uchar)(unsafe.Pointer(C.CString(s)))
	key := C.janet_keyword(cstr, C.int(len(s)))
	return C.janet_wrap_keyword(key)
}

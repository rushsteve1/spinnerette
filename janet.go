package main

/*
#cgo CFLAGS: -fPIC -O2
#cgo CFLAGS: -I ${SRCDIR}/deps/janet/build
#cgo LDFLAGS: -lm -ldl -lrt -lpthread ${SRCDIR}/deps/janet/build/libjanet.a
#include "janet.h"
Janet loader_shim(int32_t argc, Janet *argv);
const JanetReg cfuns[] = {
   {"spin/module-loader", loader_shim, "(spin/module-loader)\n\nLoads modules from Spinnerette"},
   {NULL, NULL, NULL}
};
*/
import "C"
import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
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
	C.janet_cfuns(env, C.CString(""), (*C.JanetReg)(unsafe.Pointer(&C.cfuns)))
	InitModules(env)
	return env
}

func RequestEnv(r *http.Request) (*C.JanetTable, error) {
	env := SpinEnv()

	req, err := RequestToJanet(r)
	if err != nil {
		return env, err
	}

	bindToEnv(env, "spin/request", req, "HTTP request recieved by Spinnerette")
	return env, nil
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
	buf := C.janet_pretty((*C.JanetBuffer)(C.NULL), 5, C.JANET_PRETTY_NOTRUNC, j)
	return C.GoStringN((*C.char)(unsafe.Pointer(buf.data)), buf.count)
}

func WriteResponse(j C.Janet, w http.ResponseWriter) {
	switch C.janet_type(j) {
	case C.JANET_STRING:
		w.Write([]byte(ToString(j)))
	case C.JANET_STRUCT:
		ResponseFromJanet(w, C.janet_struct_to_table(C.janet_unwrap_struct(j)))
	case C.JANET_TABLE:
		ResponseFromJanet(w, C.janet_unwrap_table(j))
	default:
		http.Error(w, "Script did not return a string or response object", 500)
	}
}

// This mirrors the structure of the Request object provided by Circlet
// https://github.com/janet-lang/circlet
func RequestToJanet(r *http.Request) (C.Janet, error) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return C.janet_wrap_nil(), err
	}

	table := C.janet_table(C.int(1024 + len(body)))

	if len(body) > 0 {
		C.janet_table_put(table, jkey("body"), jbuf(body))
	} else {
		C.janet_table_put(table, jkey("body"), C.janet_wrap_nil())
	}

	C.janet_table_put(table, jkey("uri"), jstr(r.URL.String()))
	C.janet_table_put(table, jkey("method"), jstr(r.Method))
	C.janet_table_put(table, jkey("protocol"), jstr(r.Proto))
	C.janet_table_put(table, jkey("query-string"), jstr(r.URL.RawQuery))

	headers := C.janet_table(512)
	for k, _ := range r.Header {
		C.janet_table_put(headers, jstr(k), jstr(r.Header.Get(k)))
	}
	C.janet_table_put(table, jkey("headers"), C.janet_wrap_table(headers))

	return C.janet_wrap_table(table), nil
}

func ResponseFromJanet(w http.ResponseWriter, table *C.JanetTable) {
	headers := C.janet_table_get(table, jkey("headers"))
	// TODO there is definitely a better way to do this
	// Some way to handle "dictionary" types
	if C.janet_checktype(headers, C.JANET_STRUCT) > 0 {
		headers = C.janet_wrap_table(C.janet_struct_to_table(C.janet_unwrap_struct(headers)))
	}
	if C.janet_checktype(headers, C.JANET_TABLE) > 0 {
		h := C.janet_unwrap_table(headers)
		kv := &*h.data
		for kv != (*C.JanetKV)(C.NULL) {
			k := ToString(kv.key)
			v := ToString(kv.value)
			w.Header().Add(k, v)
			kv = C.janet_dictionary_next(h.data, h.capacity, kv)
		}
	} else {
		log.Printf(":headers was not a table or struct and will not be used")
	}

	status := C.janet_table_get(table, jkey("status"))
	if C.janet_checktype(status, C.JANET_NUMBER) > 0 {
		w.WriteHeader(int(C.janet_unwrap_integer(status)))
	} else {
		http.Error(w, ":status key was not a number", 500)
	}

	body := C.janet_table_get(table, jkey("body"))
	if C.janet_checktypes(body, C.JANET_TFLAG_BYTES) > 0 {
		w.Write([]byte(ToString(body)))
	} else {
		log.Printf(":body key was not a string or buffer and will not be used")
	}
}

func bindToEnv(env *C.JanetTable, key string, value C.Janet, doc string) {
	table := C.janet_table(512)
	C.janet_table_put(table, jkey("doc"), jstr(doc))
	C.janet_table_put(table, jkey("source-map"), jstr("janet.go"))
	C.janet_table_put(table, jkey("value"), value)

	C.janet_table_put(env, jsym(key), C.janet_wrap_table(table))
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

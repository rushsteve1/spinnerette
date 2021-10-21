package bindings

// #cgo CFLAGS: -fPIC -O2
// #cgo CFLAGS: -I ${SRCDIR}/janet/build
// #cgo LDFLAGS: ${SRCDIR}/janet/build/libjanet.a ${SRCDIR}/libsqlite3.a -L. -lm -ldl -lpthread
// #include "janet.h"
import "C"
import (
	"fmt"
	"log"
	"net/http"
	"os"
	"unsafe"
)

var topEnv *C.JanetTable

func StartJanet() {
	C.janet_init()
	topEnv = initModules()
}

func StopJanet() {
	log.Printf("Stopping Janet interpreter routine")
	C.janet_deinit()
}

func EvalBytes(code []byte, source string, req *http.Request) (C.Janet, error) {
	// Ignore messages with no code
	if len(code) == 0 {
		return jnil(), nil
	}

	env := C.janet_table(0)
	env.proto = topEnv

	if req != nil {
		req, err := RequestToJanet(req)
		if err != nil {
			return jnil(), err
		}

		bindToEnv(env, "*request*", req, "Circlet-style HTTP request recieved by Spinnerette.")
	}

	var out C.Janet
	errno := C.janet_dobytes(
		env,
		(*C.uchar)(unsafe.Pointer(&code[0])),
		C.int(len(code)),
		C.CString(source),
		&out,
	)

	if errno != 0 {
		return jnil(), fmt.Errorf("janet error. number: %d", errno)
	} else {
		return out, nil
	}
}

func EvalFilePath(path string, req *http.Request) (C.Janet, error) {
	code, err := os.ReadFile(path)
	if err != nil {
		return jnil(), err
	}

	out, err := EvalBytes(code, path, req)
	if err != nil {
		return jnil(), err
	}

	return out, nil
}

func EvalString(code string) (C.Janet, error) {
	return EvalBytes([]byte(code), "Spinnerette Internal", nil)
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
	v := C.janet_table_get(env, jsym(key))
	if C.janet_checktype(v, C.JANET_NIL) > 0 {
		log.Fatal("Could not get env key: ", key)
	}
	table := C.janet_unwrap_table(v)
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

func jnil() C.Janet {
	return C.janet_wrap_nil()
}

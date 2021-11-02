package bindings

// #include "janet.h"
import "C"
import (
	"unsafe"
)

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

func envResolve(env *C.JanetTable, name string) C.Janet {
	var out C.Janet
	C.janet_resolve(env, C.janet_symbol(uchars(name), C.int(len(name))), &out)
	return out
}

func jbuf(b []byte) C.Janet {
	buf := C.janet_buffer(C.int(len(b)))
	C.janet_buffer_push_bytes(buf, (*C.uchar)(unsafe.Pointer(&b[0])), C.int(len(b)))
	return C.janet_wrap_buffer(buf)
}

func uchars(s string) *C.uchar {
	return (*C.uchar)(unsafe.Pointer(C.CString(s)))
}

// These could possibly be replaced by the C macro versions that Janet provides
func jstr(s string) C.Janet {
	return C.janet_wrap_string(C.janet_string(uchars(s), C.int(len(s))))
}

func jsym(s string) C.Janet {
	return C.janet_wrap_symbol(C.janet_symbol(uchars(s), C.int(len(s))))
}

func jkey(s string) C.Janet {
	return C.janet_wrap_keyword(C.janet_keyword(uchars(s), C.int(len(s))))
}

func jnil() C.Janet {
	return C.janet_wrap_nil()
}

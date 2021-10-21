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
	"runtime"
	"unsafe"
)

type message struct {
	Code   []byte
	Source string
	Req    *http.Request
}

var sendChan chan message
var recvChan chan C.Janet
var errChan chan error

func StartJanet() {
	sendChan = make(chan message)
	recvChan = make(chan C.Janet)
	errChan = make(chan error)
	go janetRoutine()
}

func StopJanet() {
	log.Printf("Stopping Janet interpreter routine")
	close(sendChan)
}

func janetRoutine() {
	defer close(recvChan)

	// The Janet interpreter is not thread-safe
	// So we give it it's own entire thread
	// However this means we need at least 2 threads
	c := runtime.NumCPU()
	if c < 2 {
		c = 2
	}
	runtime.GOMAXPROCS(c)
	runtime.LockOSThread()

	C.janet_init()
	defer C.janet_deinit()

	topEnv := initModules()

	for msg := range sendChan {
		// Ignore messages with no code
		if len(msg.Code) == 0 {
			continue
		}

		env := C.janet_table(0)
		env.proto = topEnv

		if msg.Req != nil {
			req, err := RequestToJanet(msg.Req)
			if err != nil {
				errChan <- err
				continue
			}

			bindToEnv(env, "*request*", req, "Circlet-style HTTP request recieved by Spinnerette.")
		}

		out, err := innerEval(env, msg.Code, msg.Source)

		if err != nil {
			errChan <- err
		} else {
			recvChan <- out
		}
		
	}
}

func innerEval(env *C.JanetTable, code []byte, source string) (C.Janet, error) {
	// TODO this locks up on the second evaluation
	var out C.Janet
	errno := C.janet_dobytes(
		env,
		(*C.uchar)(unsafe.Pointer(&code[0])),
		C.int(len(code)),
		C.CString(source),
		&out,
	)

	if errno != 0 {
		return jnil(), fmt.Errorf("janet error number: %d", errno)
	} 
	return out, nil
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

func EvalBytes(code []byte, source string, req *http.Request) (C.Janet, error) {
	msg := message{
		Code:   code,
		Source: source,
		Req:    req,
	}

	sendChan <- msg
	select {
	case out := <-recvChan:
		return out, nil
	case err := <-errChan:
		return jnil(), err
	}
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

func jnil() C.Janet {
	return C.janet_wrap_nil()
}

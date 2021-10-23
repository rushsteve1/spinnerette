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

type evalMsg struct {
	Code   []byte
	Source string
	Req    *http.Request
}

type writeMsg struct {
	J C.Janet
	W http.ResponseWriter
}

var quitChan chan bool
var evalChan chan evalMsg
var writeChan chan writeMsg
var recvChan chan C.Janet
var errChan chan error

func StartJanet() {
	quitChan = make(chan bool)
	evalChan = make(chan evalMsg)
	writeChan = make(chan writeMsg)
	recvChan = make(chan C.Janet)
	errChan = make(chan error)
	go janetRoutine()
}

func StopJanet() {
	close(quitChan)
}

func janetRoutine() {
	defer close(recvChan)

	// The Janet interpreter is not thread-safe
	// So we give it it's own entire thread
	runtime.LockOSThread()

	C.janet_init()
	defer C.janet_deinit()

	topEnv := initModules()

	for {
		select {
		case <-quitChan:
			log.Printf("Stopping Janet interpreter routine")
			return
		case msg := <-evalChan:
			evalRoutine(msg, topEnv)
		case msg := <-writeChan:
			writeResponse(msg.J, msg.W)
		}
	}
}

func evalRoutine(msg evalMsg, topEnv *C.JanetTable) {
	// Ignore messages with no code
	if len(msg.Code) == 0 {
		return
	}

	env := C.janet_table(0)
	env.proto = topEnv

	if msg.Req != nil {
		req, err := requestToJanet(msg.Req)
		if err != nil {
			errChan <- err
			return
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

func innerEval(env *C.JanetTable, code []byte, source string) (C.Janet, error) {
	if len(code) == 0 {
		return jnil(), fmt.Errorf("No code given to eval for source: %s", source)
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
	msg := evalMsg{
		Code:   code,
		Source: source,
		Req:    req,
	}

	evalChan <- msg
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

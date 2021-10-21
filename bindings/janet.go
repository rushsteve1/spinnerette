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
	
	// Block now to ensure everything is started
	sendChan <- message { Code: []byte{}, Source: "", Req: nil }
}

func StopJanet() {
	log.Printf("Stopping Janet interpreter routine")
	close(sendChan)
}

func janetRoutine() {
	defer close(recvChan)

	C.janet_init()
	defer C.janet_deinit()

	topEnv := C.janet_core_env((*C.JanetTable)(C.NULL))
	initModules(topEnv)

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

		var out C.Janet
		errno := C.janet_dobytes(
			env,
			(*C.uchar)(unsafe.Pointer(&msg.Code[0])),
			C.int(len(msg.Code)),
			C.CString(msg.Source),
			&out,
		)
	
		if errno != 0 {
			errChan <- fmt.Errorf("Janet error. number: %d", errno)
		} else {
			recvChan <- out
		}
	}

	// In theory this should never happen
	// But those are famous last words, so log it if it does happen
	log.Fatal("Janet Message Loop Ended")
}

func EvalFilePath(path string, req *http.Request) (*C.Janet, error) {
	code, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	out, err := EvalBytes(code, path, req)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func EvalBytes(code []byte, source string, req *http.Request) (*C.Janet, error) {
	msg := message{
		Code:   code,
		Source: source,
		Req:    req,
	}

	sendChan <- msg
	select {
	case out := <-recvChan:
		return &out, nil
	case err := <-errChan:
		return nil, err
	}
}

func EvalString(code string) (*C.Janet, error) {
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

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

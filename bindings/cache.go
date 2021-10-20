package bindings

// #include "janet.h"
import "C"
import (
	"time"
)

type CacheValue struct {
	Value C.Janet
	At    float64
}

var SpinCache = map[string]CacheValue{}

//export CacheGet
func CacheGet(k *C.char) C.Janet {
	key := C.GoString(k)
	var val C.Janet
	var at float64
	if v, ok := SpinCache[key]; ok {
		val = v.Value
		at = v.At
	} else {
		val = C.janet_wrap_nil()
		at = -1.0
	}

	// Janet only supports 32-bit integers, so we have to use a double
	tup := []C.Janet{val, C.janet_wrap_number(C.double(at))}
	return C.janet_wrap_tuple(C.janet_tuple_n(&tup[0], 2))
}

//export CacheSet
func CacheSet(k *C.char, value C.Janet) C.Janet {
	key := C.GoString(k)
	// TODO since these values have to live across multiple invocations of the
	// janet runtime, things might get tricky
	if v, ok := SpinCache[key]; ok {
		C.janet_gcunroot(v.Value)
	}
	C.janet_gcroot(value)
	SpinCache[key] = CacheValue{Value: value, At: float64(time.Now().Unix())}
	return value
}

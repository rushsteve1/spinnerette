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
	SpinCache[key] = CacheValue{Value: value, At: float64(time.Now().Unix())}
	return value
}

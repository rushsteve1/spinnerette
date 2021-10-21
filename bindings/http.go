package bindings

// #include "janet.h"
import "C"
import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func WriteResponse(j C.Janet, w http.ResponseWriter) {
	switch C.janet_type(j) {
	case C.JANET_BUFFER:
		fallthrough
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
		return jnil(), err
	}

	table := C.janet_table(C.int(1024 + len(body)))

	if len(body) > 0 {
		C.janet_table_put(table, jkey("body"), jbuf(body))
	} else {
		C.janet_table_put(table, jkey("body"), jnil())
	}

	C.janet_table_put(table, jkey("uri"), jstr(r.URL.String()))
	C.janet_table_put(table, jkey("method"), jstr(r.Method))
	C.janet_table_put(table, jkey("protocol"), jstr(r.Proto))
	C.janet_table_put(table, jkey("query-string"), jstr(r.URL.RawQuery))

	headers := C.janet_table(512)
	for k := range r.Header {
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
		kv := h.data
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

func RenderTemple(path string, req *http.Request) (C.Janet, error) {
	code, err := os.ReadFile(path)
	if err != nil {
		return jnil(), err
	}

	escapedCode := strings.ReplaceAll(string(code), "\"", "\\\"")
	path = strings.ReplaceAll(path, "\"", "\\\"")

	fn := fmt.Sprintf("(import spork/temple :as temple) (let [out @\"\"] (with-dyns [:out out] ((temple/create \"%s\" \"%s\") {})) out)", escapedCode, path)
	out, err := EvalBytes([]byte(fn), path, req)
	if err != nil {
		return jnil(), err
	}
	return out, nil
}

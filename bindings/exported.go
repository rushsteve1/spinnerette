package bindings

// #include "janet.h"
import "C"
import (
	"html"
	"unsafe"

	"github.com/russross/blackfriday/v2"
)

//export renderMarkdown
func renderMarkdown(str unsafe.Pointer, len C.int) *C.uchar {
	md := blackfriday.Run(C.GoBytes(str, len))
	return uchars(string(md))
}

var params blackfriday.HTMLRendererParameters = blackfriday.HTMLRendererParameters{
	Flags: blackfriday.HTMLFlagsNone,
}
var rend = blackfriday.NewHTMLRenderer(params)

//export renderMarkdownUnescaped
func renderMarkdownUnescaped(str unsafe.Pointer, len C.int) *C.uchar {
	md := blackfriday.Run(C.GoBytes(str, len), blackfriday.WithRenderer(rend))
	mds := html.UnescapeString(string(md))
	return uchars(mds)
}

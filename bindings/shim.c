/*
 * This file contains "shim" functions which exist to wrap Go functions so that
 * they can be bound into Janet cfunctions
 * It is automatically included at compile-time by cgo
 */

#include "./shared.h"
#include "_cgo_export.h"

JANET_FN_SD(deep_pretty, "(spinternal/deep-pretty x)", "Returns a pretty string of maximum depth.") {
   janet_arity(argc, 0, -1);
   JanetBuffer* buf = janet_buffer(128);
   for (unsigned int i; i < argc; i++) {
      janet_pretty(buf, -1, JANET_PRETTY_NOTRUNC, argv[i]);
   }
   return janet_wrap_buffer(buf);
}

JANET_FN_SD(markdown_shim, "(spinternal/markdown str)", "Renders a Markdown string to HTML.") {
   janet_fixarity(argc, 1);
   char* s = (char*)janet_unwrap_string(argv[0]);
   uint8_t* md = renderMarkdown(s, strlen(s));
   return janet_wrap_string(md);
}

JANET_FN_SD(markdown_unescaped_shim, "(spinternal/markdown-unescaped str)", "Renders a Markdown string to HTML but with values unescaped. Suitable for passing to Temple.") {
   janet_fixarity(argc, 1);
   char* s = (char*)janet_unwrap_string(argv[0]);
   uint8_t* md = renderMarkdownUnescaped(s, strlen(s));
   return janet_wrap_string(md);
}

const JanetRegExt spin_cfuns[] = {
   // TODO use JANET_REG_SD
   JANET_REG_("deep-pretty", deep_pretty),
   JANET_REG_("markdown", markdown_shim),
   JANET_REG_("markdown-unescaped", markdown_unescaped_shim),
   JANET_REG_END
};

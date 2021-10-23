/*
 * This file contains "shim" functions which exist to wrap Go functions so that
 * they can be bound into Janet cfunctions
 * It is automatically included at compile-time by cgo
 */

#include "./shared.h"
#include "_cgo_export.h"

JANET_FN_SD(deep_pretty, "(deep-pretty x)", "Returns a pretty string of maximum depth") {
   janet_arity(argc, 0, -1);
   JanetBuffer* buf = janet_buffer(128);
   for (unsigned int i; i < argc; i++) {
      janet_pretty(buf, -1, JANET_PRETTY_NOTRUNC, argv[i]);
   }
   return janet_wrap_buffer(buf);
}

const JanetRegExt spin_cfuns[] = {
   // TODO use JANET_REG_SD
   JANET_REG_("deep-pretty", deep_pretty),
   JANET_REG_END
};

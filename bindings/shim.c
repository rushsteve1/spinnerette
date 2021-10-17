/*
 * This file contains "shim" functions which exist to wrap Go functions so that
 * they can be bound into Janet cfunctions
 * It is automatically included at compile-time by cgo
 */

#ifndef SHIM_C_
#define SHIM_C_

#include "janet.h"
#include "_cgo_export.h"

Janet loader_shim(int32_t argc, Janet *argv) {
  janet_arity(argc, 1, 2);
  char* path = (char*) janet_getcstring(argv, 0);
  return janet_wrap_table(ModuleLoader(path));
}

const JanetReg shim_cfuns[] = {
   {"spin/module-loader", loader_shim,
    "(spin/module-loader)\n\nLoader for embedded Spinnerette modules"
   },
   {NULL, NULL, NULL}
};

#endif // SHIM_C_

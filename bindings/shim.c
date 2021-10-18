/*
 * This file contains "shim" functions which exist to wrap Go functions so that
 * they can be bound into Janet cfunctions
 * It is automatically included at compile-time by cgo
 */

#include "./shared.h"
#include "_cgo_export.h"

// TODO make these static?
Janet loader_shim(int32_t argc, Janet *argv) {
  janet_arity(argc, 1, 2);
  char* path = (char*) janet_getcstring(argv, 0);
  return janet_wrap_table(ModuleLoader(path, janet_current_fiber()->env));
}

Janet pretty(int32_t argc, Janet *argv) {
  janet_fixarity(argc, 1);
  return janet_wrap_buffer(janet_pretty(NULL, 5, JANET_PRETTY_NOTRUNC, argv[0]));
}

Janet path_pred_shim(int32_t argc, Janet *argv) {
  janet_fixarity(argc, 1);
  janet_getcstring(argv, 0); // just make sure it's a string
  return PathPred(argv[0]);
}

const JanetReg shim_cfuns[] = {
   {"spin/module-loader", loader_shim,
       "(spin/module-loader)\n\nLoader for embedded Spinnerette modules."
   },
   {"spin/path-pred", path_pred_shim,
       "(spin/path-pred)\n\nPredicate that verifies built-in paths."},
   {"pretty", pretty,
       "(pretty x)\n\nReturns the non-truncated pretty string."
   },
   {NULL, NULL, NULL}
};

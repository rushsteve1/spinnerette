/*
 * This file contains "shim" functions which exist to wrap Go functions so that
 * they can be bound into Janet cfunctions
 * It is automatically included at compile-time by cgo
 */

#include "./shared.h"
#include "_cgo_export.h"

JANET_CFUN(loader_shim) {
  janet_arity(argc, 1, 2);
  char* path = (char*) janet_getcstring(argv, 0);
  return janet_wrap_table(moduleLoader(path, janet_current_fiber()->env));
}

JANET_CFUN(path_pred_shim) {
  janet_fixarity(argc, 1);
  janet_getcstring(argv, 0); // just make sure it's a string
  return pathPred(argv[0]);
}

const JanetReg spin_cfuns[] = {
   {"module-loader", loader_shim,
       "(spin/module-loader x &args)\n\nLoader for embedded Spinnerette modules."
   },
   {"path-pred", path_pred_shim,
       "(spinternal/path-pred x)\n\nPredicate that verifies and expands import"
       "paths for bundled libraries."},
   {NULL, NULL, NULL}
};

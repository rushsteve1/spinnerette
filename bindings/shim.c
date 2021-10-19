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

Janet cache_get_shim(int32_t argc, Janet *argv) {
  janet_fixarity(argc, 1);
  char* key = (char*)janet_getkeyword(argv, 0);
  return CacheGet(key);
}

Janet cache_set_shim(int32_t argc, Janet *argv) {
  janet_fixarity(argc, 2);
  char* key = (char*)janet_getkeyword(argv, 0);
  Janet value = argv[1];
  return CacheSet(key, value);
}

const JanetReg shim_cfuns[] = {
   {"spin/module-loader", loader_shim,
       "(spin/module-loader x &args)\n\nLoader for embedded Spinnerette modules."
   },
   {"spin/path-pred", path_pred_shim,
       "(spin/path-pred x)\n\nPredicate that verifies built-in paths."},
   {"spin/cache-get", cache_get_shim,
       "(spin/cache-get key)\n\nGets a value from the Spinnerette cache."
       "Returns a tuple of the value and the UNIX time that it was cached."
       "If the key was not in the cache returns (nil -1)"},
   {"spin/cache-set", cache_set_shim,
       "(spin/cache-set key value)\n\nSets a value in the Spinnerette cache."
       "Returns the given `value`."},
   {"pretty", pretty,
       "(pretty x)\n\nReturns the non-truncated pretty string."
   },
   {NULL, NULL, NULL}
};

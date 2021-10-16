#include "janet.h"
#include "_cgo_export.h"

Janet loader_shim(int32_t argc, Janet *argv) {
  janet_arity(argc, 1, -1);
  char* path = (char*) janet_getcstring(argv, 0);
  // JanetTable* args = janet_opttable(argv, argc, 1, 0);

  return janet_wrap_table(ModuleLoader(path));
}

#ifndef SHARED_H_
#define SHARED_H_

#define CONCAT(a, b) CONCAT_INNER(a, b)
#define CONCAT_INNER(a, b) a ## b
#define JANET_ENTRY_NAME CONCAT(_janet_init, __LINE__)

#include "janet.h"

typedef struct { JanetReg *cfuns; } _cfun_ns;

extern const JanetRegExt spin_cfuns[];
extern _cfun_ns const sqlite3_ns;
extern _cfun_ns const json_ns;

#endif // SHARED_H_

/*
 * dlfcn.h - Dynamic Loading Interface for GOC
 * 
 * This header defines the dlopen/dlsym/dlclose/dlerror interface
 * for dynamic library loading. Lua uses this for require() and
 * dynamic module loading.
 * 
 * Implementation: Full ELF64 shared library loader
 */

#ifndef _DLFCN_H
#define _DLFCN_H

/*
 * dlopen flags
 * 
 * RTLD_LAZY:  Resolve symbols on first use (lazy binding)
 * RTLD_NOW:   Resolve all symbols immediately (eager binding)
 * RTLD_LOCAL: Symbols not visible to other libraries (default)
 * RTLD_GLOBAL: Symbols visible to other libraries
 */
#define RTLD_LAZY   1
#define RTLD_NOW    2
#define RTLD_LOCAL  0
#define RTLD_GLOBAL 4

/*
 * dlopen - Load dynamic library
 * 
 * @filename: Path to shared library (.so file)
 * @flag:     Combination of RTLD_* flags
 * @returns:  Opaque handle on success, NULL on failure
 * 
 * Loads the specified shared library and returns a handle.
 * The handle is used in subsequent dlsym and dlclose calls.
 * 
 * If filename is NULL, returns a handle for the main program.
 * 
 * On failure, returns NULL and dlerror() will return error message.
 */
void *dlopen(const char *filename, int flag);

/*
 * dlsym - Look up symbol in dynamic library
 * 
 * @handle: Handle from dlopen
 * @symbol: Name of symbol to find
 * @returns: Address of symbol, NULL if not found
 * 
 * Searches for the named symbol in the loaded library.
 * Returns the address of the symbol, which can be cast to
 * appropriate function or data pointer type.
 * 
 * On failure, returns NULL and dlerror() will return error message.
 */
void *dlsym(void *handle, const char *symbol);

/*
 * dlclose - Unload dynamic library
 * 
 * @handle: Handle from dlopen
 * @returns: 0 on success, non-zero on failure
 * 
 * Decrements the reference count of the library.
 * When reference count reaches 0, the library is unloaded.
 * 
 * Note: After dlclose, handles and symbol addresses from this
 * library become invalid.
 */
int dlclose(void *handle);

/*
 * dlerror - Get error message
 * 
 * @returns: Error message string, or NULL if no error
 * 
 * Returns a human-readable error message describing the last
 * error that occurred. Returns NULL if no error has occurred
 * since the last dlerror call.
 * 
 * Each call to dlerror clears the error state, so calling
 * dlerror twice will return NULL the second time (if no new error).
 */
char *dlerror(void);

/*
 * Implementation Notes:
 * 
 * 1. ELF64 Shared Library Format:
 *    - ELF header identifies as ET_DYN (shared object)
 *    - Program headers specify loadable segments
 *    - Dynamic section contains symbol tables and relocations
 * 
 * 2. Loading Process (dlopen):
 *    a. Open and validate ELF file
 *    b. Parse ELF header and program headers
 *    c. Allocate memory for segments
 *    d. Load segments into memory
 *    e. Process relocations
 *    f. Call initialization functions (.init array)
 *    g. Return handle
 * 
 * 3. Symbol Resolution (dlsym):
 *    a. Find symbol table (.dynsym or .symtab)
 *    b. Find string table (.dynstr or .strtab)
 *    c. Search for symbol name
 *    d. Return symbol address (with relocation applied)
 * 
 * 4. Handle Structure (internal):
 *    typedef struct {
 *        void *base;           // Base address of loaded library
 *        char *path;           // Library path
 *        int refcount;         // Reference count
 *        void *symtab;         // Symbol table
 *        void *strtab;         // String table
 *        // ... more fields
 *    } dl_handle_t;
 * 
 * 5. Error Handling:
 *    - Store last error in thread-local or global buffer
 *    - dlerror() returns this buffer
 *    - Clear buffer on each dlerror() call
 */

#endif /* _DLFCN_H */
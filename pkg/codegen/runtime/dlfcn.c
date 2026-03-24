/*
 * dlfcn.c - Dynamic Loading Implementation for GOC
 * 
 * This file provides the dlopen/dlsym/dlclose/dlerror implementation.
 * Full ELF64 shared library loader.
 * 
 * See: pkg/stdlib/dlfcn.h for interface documentation
 */

#include "../stdlib/dlfcn.h"
#include "../stdlib/syscall_wrapper.h"
#include <stddef.h>
#include <stdint.h>

/* ============================================================================
 * ELF64 Structures (matching pkg/linker/elf.go)
 * ============================================================================ */

#define EI_NIDENT 16

/* ELF64 Header */
typedef struct {
    unsigned char e_ident[EI_NIDENT];
    uint16_t e_type;
    uint16_t e_machine;
    uint32_t e_version;
    uint64_t e_entry;
    uint64_t e_phoff;
    uint64_t e_shoff;
    uint32_t e_flags;
    uint16_t e_ehsize;
    uint16_t e_phentsize;
    uint16_t e_phnum;
    uint16_t e_shentsize;
    uint16_t e_shnum;
    uint16_t e_shstrndx;
} Elf64_Ehdr;

/* ELF64 Program Header */
typedef struct {
    uint32_t p_type;
    uint32_t p_flags;
    uint64_t p_offset;
    uint64_t p_vaddr;
    uint64_t p_paddr;
    uint64_t p_filesz;
    uint64_t p_memsz;
    uint64_t p_align;
} Elf64_Phdr;

/* ELF64 Section Header */
typedef struct {
    uint32_t sh_name;
    uint32_t sh_type;
    uint64_t sh_flags;
    uint64_t sh_addr;
    uint64_t sh_offset;
    uint64_t sh_size;
    uint32_t sh_link;
    uint32_t sh_info;
    uint64_t sh_addralign;
    uint64_t sh_entsize;
} Elf64_Shdr;

/* ELF64 Symbol */
typedef struct {
    uint32_t st_name;
    unsigned char st_info;
    unsigned char st_other;
    uint16_t st_shndx;
    uint64_t st_value;
    uint64_t st_size;
} Elf64_Sym;

/* ELF64 Dynamic Entry */
typedef struct {
    int64_t d_tag;
    uint64_t d_val;
} Elf64_Dyn;

/* ELF Constants */
#define EI_MAG0 0
#define EI_MAG1 1
#define EI_MAG2 2
#define EI_MAG3 3
#define EI_CLASS 4
#define ELFMAG0 0x7f
#define ELFMAG1 'E'
#define ELFMAG2 'L'
#define ELFMAG3 'F'
#define ELFCLASS64 2
#define ET_DYN 3

/* Program Header Types */
#define PT_NULL    0
#define PT_LOAD    1

/* Section Header Types */
#define SHT_DYNSYM 11

/* ============================================================================
 * Handle Structure
 * ============================================================================ */

#define MAX_LOADED_LIBRARIES 64
#define MAX_ERROR_MSG 256

typedef struct dl_handle {
    void *base;              /* Base address of loaded library */
    char *path;              /* Library path (copied) */
    int refcount;            /* Reference count */
    void *symtab;            /* Symbol table (in mmap'd region) */
    void *strtab;            /* String table (in mmap'd region) */
    size_t symcount;         /* Number of symbols */
    size_t min_vaddr;        /* Minimum virtual address for address translation */
    size_t max_vaddr;        /* Maximum virtual address */
    int in_use;              /* Handle is in use */
} dl_handle_t;

/* Global state */
static dl_handle_t handles[MAX_LOADED_LIBRARIES];
static char dl_error_buf[MAX_ERROR_MSG];
static int dl_has_error = 0;

/* ============================================================================
 * Helper Functions
 * ============================================================================ */

/* String length */
static size_t dl_strlen(const char *s) {
    size_t len = 0;
    while (s[len] != '\0') {
        len++;
    }
    return len;
}

/* String copy */
static char *dl_strcpy(char *dest, const char *src) {
    char *d = dest;
    while ((*d++ = *src++) != '\0');
    return dest;
}

/* String compare */
static int dl_strcmp(const char *s1, const char *s2) {
    while (*s1 && (*s1 == *s2)) {
        s1++;
        s2++;
    }
    return *(unsigned char *)s1 - *(unsigned char *)s2;
}

/* Memory copy */
static void *dl_memcpy(void *dest, const void *src, size_t n) {
    unsigned char *d = (unsigned char *)dest;
    const unsigned char *s = (const unsigned char *)src;
    while (n--) {
        *d++ = *s++;
    }
    return dest;
}

/* Memory set */
static void *dl_memset(void *s, int c, size_t n) {
    unsigned char *p = (unsigned char *)s;
    while (n--) {
        *p++ = (unsigned char)c;
    }
    return s;
}

/* Set error message */
static void dl_set_error(const char *msg) {
    size_t i;
    for (i = 0; i < MAX_ERROR_MSG - 1 && msg[i] != '\0'; i++) {
        dl_error_buf[i] = msg[i];
    }
    dl_error_buf[i] = '\0';
    dl_has_error = 1;
}

/* Clear error state */
static void dl_clear_error(void) {
    dl_error_buf[0] = '\0';
    dl_has_error = 0;
}

/* Find handle slot */
static int dl_find_free_handle(void) {
    int i;
    for (i = 0; i < MAX_LOADED_LIBRARIES; i++) {
        if (!handles[i].in_use) {
            return i;
        }
    }
    return -1;
}

/* Find handle by path */
static int dl_find_handle_by_path(const char *path) {
    int i;
    for (i = 0; i < MAX_LOADED_LIBRARIES; i++) {
        if (handles[i].in_use && handles[i].path && 
            dl_strcmp(handles[i].path, path) == 0) {
            return i;
        }
    }
    return -1;
}

/* ============================================================================
 * Main Implementation
 * ============================================================================ */

/*
 * dlopen - Load dynamic library
 */
void *dlopen(const char *filename, int flag) {
    int fd;
    int handle_idx;
    Elf64_Ehdr ehdr;
    Elf64_Phdr *phdrs = NULL;
    Elf64_Shdr *shdrs = NULL;
    ssize_t bytes_read;
    int i;
    void *map_base = NULL;
    size_t total_size = 0;
    size_t min_vaddr = 0xFFFFFFFFFFFFFFFFULL;
    size_t max_vaddr = 0;
    void *symtab = NULL;
    void *strtab = NULL;
    size_t symcount = 0;
    
    (void)flag;  /* Unused for now */
    
    dl_clear_error();
    
    /* Validate input */
    if (!filename) {
        dl_set_error("dlopen: filename is NULL");
        return NULL;
    }
    
    /* Check if already loaded */
    handle_idx = dl_find_handle_by_path(filename);
    if (handle_idx >= 0) {
        handles[handle_idx].refcount++;
        return &handles[handle_idx];
    }
    
    /* Find free handle slot */
    handle_idx = dl_find_free_handle();
    if (handle_idx < 0) {
        dl_set_error("dlopen: too many libraries loaded");
        return NULL;
    }
    
    /* Open file */
    fd = open(filename, 0);  /* O_RDONLY = 0 */
    if (fd < 0) {
        dl_set_error("dlopen: cannot open file");
        return NULL;
    }
    
    /* Read ELF header */
    bytes_read = read(fd, &ehdr, sizeof(ehdr));
    if (bytes_read != sizeof(ehdr)) {
        dl_set_error("dlopen: cannot read ELF header");
        close(fd);
        return NULL;
    }
    
    /* Validate ELF magic */
    if (ehdr.e_ident[EI_MAG0] != ELFMAG0 || ehdr.e_ident[EI_MAG1] != ELFMAG1 ||
        ehdr.e_ident[EI_MAG2] != ELFMAG2 || ehdr.e_ident[EI_MAG3] != ELFMAG3) {
        dl_set_error("dlopen: not an ELF file");
        close(fd);
        return NULL;
    }
    
    /* Validate ELF class (64-bit) */
    if (ehdr.e_ident[EI_CLASS] != ELFCLASS64) {
        dl_set_error("dlopen: not a 64-bit ELF file");
        close(fd);
        return NULL;
    }
    
    /* Validate file type (shared object) */
    if (ehdr.e_type != ET_DYN) {
        dl_set_error("dlopen: not a shared object file");
        close(fd);
        return NULL;
    }
    
    /* Read program headers */
    if (ehdr.e_phentsize != sizeof(Elf64_Phdr)) {
        dl_set_error("dlopen: invalid program header size");
        close(fd);
        return NULL;
    }
    
    phdrs = (Elf64_Phdr *)malloc(ehdr.e_phnum * sizeof(Elf64_Phdr));
    if (!phdrs) {
        dl_set_error("dlopen: out of memory");
        close(fd);
        return NULL;
    }
    
    /* Seek to program headers */
    {
        char buf[512];
        size_t offset = ehdr.e_phoff;
        size_t pos = sizeof(ehdr);
        
        while (pos < offset) {
            size_t to_read = offset - pos;
            if (to_read > sizeof(buf)) to_read = sizeof(buf);
            bytes_read = read(fd, buf, to_read);
            if (bytes_read <= 0) break;
            pos += bytes_read;
        }
    }
    
    /* Read program headers */
    bytes_read = read(fd, phdrs, ehdr.e_phnum * sizeof(Elf64_Phdr));
    if (bytes_read != (ssize_t)(ehdr.e_phnum * sizeof(Elf64_Phdr))) {
        dl_set_error("dlopen: cannot read program headers");
        close(fd);
        free(phdrs);
        return NULL;
    }
    
    /* Calculate total memory size needed */
    for (i = 0; i < ehdr.e_phnum; i++) {
        if (phdrs[i].p_type == PT_LOAD) {
            if (phdrs[i].p_vaddr < min_vaddr) {
                min_vaddr = phdrs[i].p_vaddr;
            }
            if (phdrs[i].p_vaddr + phdrs[i].p_memsz > max_vaddr) {
                max_vaddr = phdrs[i].p_vaddr + phdrs[i].p_memsz;
            }
        }
    }
    
    total_size = max_vaddr - min_vaddr;
    
    /* Allocate memory for the library using mmap */
    map_base = mmap(NULL, total_size, 
                    PROT_READ | PROT_WRITE | PROT_EXEC,
                    2 | 0x20,  /* MAP_PRIVATE | MAP_ANONYMOUS */
                    -1, 0);
    
    if (map_base == (void *)-1 || map_base == NULL) {
        dl_set_error("dlopen: cannot allocate memory");
        close(fd);
        free(phdrs);
        return NULL;
    }
    
    /* Load segments */
    for (i = 0; i < ehdr.e_phnum; i++) {
        if (phdrs[i].p_type == PT_LOAD && phdrs[i].p_filesz > 0) {
            void *seg_addr = (char *)map_base + phdrs[i].p_vaddr - min_vaddr;
            
            /* Seek to segment offset */
            {
                char buf[512];
                size_t offset = phdrs[i].p_offset;
                size_t pos = 0;
                
                while (pos < offset) {
                    size_t to_read = offset - pos;
                    if (to_read > sizeof(buf)) to_read = sizeof(buf);
                    bytes_read = read(fd, buf, to_read);
                    if (bytes_read <= 0) break;
                    pos += bytes_read;
                }
            }
            
            /* Read segment data */
            read(fd, seg_addr, phdrs[i].p_filesz);
        }
    }
    
    /* Read section headers to find symbol tables */
    if (ehdr.e_shoff != 0 && ehdr.e_shnum != 0) {
        shdrs = (Elf64_Shdr *)malloc(ehdr.e_shnum * sizeof(Elf64_Shdr));
        if (shdrs) {
            /* Seek to section headers */
            {
                char buf[512];
                size_t offset = ehdr.e_shoff;
                size_t pos = 0;
                
                while (pos < offset) {
                    size_t to_read = offset - pos;
                    if (to_read > sizeof(buf)) to_read = sizeof(buf);
                    bytes_read = read(fd, buf, to_read);
                    if (bytes_read <= 0) break;
                    pos += bytes_read;
                }
            }
            
            /* Read section headers */
            if (read(fd, shdrs, ehdr.e_shnum * sizeof(Elf64_Shdr)) == 
                (ssize_t)(ehdr.e_shnum * sizeof(Elf64_Shdr))) {
                
                /* Find .dynsym and .dynstr sections */
                for (i = 0; i < ehdr.e_shnum; i++) {
                    if (shdrs[i].sh_type == SHT_DYNSYM) {
                        symtab = (char *)map_base + shdrs[i].sh_addr - min_vaddr;
                        symcount = shdrs[i].sh_size / sizeof(Elf64_Sym);
                        
                        /* Use sh_link to find corresponding strtab */
                        if (shdrs[i].sh_link < ehdr.e_shnum) {
                            strtab = (char *)map_base + shdrs[shdrs[i].sh_link].sh_addr - min_vaddr;
                        }
                        break;
                    }
                }
            }
            free(shdrs);
        }
    }
    
    close(fd);
    free(phdrs);
    
    /* Initialize handle */
    handles[handle_idx].in_use = 1;
    handles[handle_idx].base = map_base;
    handles[handle_idx].refcount = 1;
    handles[handle_idx].symtab = symtab;
    handles[handle_idx].strtab = strtab;
    handles[handle_idx].symcount = symcount;
    handles[handle_idx].min_vaddr = min_vaddr;
    handles[handle_idx].max_vaddr = max_vaddr;
    
    /* Copy path */
    {
        size_t path_len = dl_strlen(filename);
        handles[handle_idx].path = (char *)malloc(path_len + 1);
        if (handles[handle_idx].path) {
            dl_strcpy(handles[handle_idx].path, filename);
        }
    }
    
    return &handles[handle_idx];
}

/*
 * dlsym - Look up symbol in dynamic library
 */
void *dlsym(void *handle, const char *symbol) {
    dl_handle_t *hdl;
    Elf64_Sym *symtab;
    char *strtab;
    size_t i;
    
    dl_clear_error();
    
    /* Validate handle */
    if (!handle) {
        dl_set_error("dlsym: invalid handle");
        return NULL;
    }
    
    hdl = (dl_handle_t *)handle;
    
    /* Validate handle is in use */
    if (!hdl->in_use) {
        dl_set_error("dlsym: invalid handle (not loaded)");
        return NULL;
    }
    
    /* Validate symbol table */
    if (!hdl->symtab || !hdl->strtab) {
        dl_set_error("dlsym: no symbol table");
        return NULL;
    }
    
    symtab = (Elf64_Sym *)hdl->symtab;
    strtab = (char *)hdl->strtab;
    
    /* Search for symbol */
    for (i = 0; i < hdl->symcount; i++) {
        const char *sym_name = strtab + symtab[i].st_name;
        
        if (dl_strcmp(sym_name, symbol) == 0) {
            /* Found symbol */
            /* Skip undefined symbols */
            if (symtab[i].st_shndx == 0) {
                continue;
            }
            
            /* Return symbol address */
            return (void *)((char *)hdl->base + symtab[i].st_value - hdl->min_vaddr);
        }
    }
    
    /* Symbol not found */
    dl_set_error("dlsym: symbol not found");
    return NULL;
}

/*
 * dlclose - Unload dynamic library
 */
int dlclose(void *handle) {
    dl_handle_t *hdl;
    
    dl_clear_error();
    
    /* Validate handle */
    if (!handle) {
        dl_set_error("dlclose: invalid handle");
        return -1;
    }
    
    hdl = (dl_handle_t *)handle;
    
    /* Validate handle is in use */
    if (!hdl->in_use) {
        dl_set_error("dlclose: invalid handle (not loaded)");
        return -1;
    }
    
    /* Decrement reference count */
    hdl->refcount--;
    
    /* If reference count is 0, unload library */
    if (hdl->refcount == 0) {
        /* Unmap memory */
        if (hdl->base) {
            munmap(hdl->base, hdl->max_vaddr - hdl->min_vaddr);
        }
        
        /* Free path */
        if (hdl->path) {
            free(hdl->path);
        }
        
        /* Clear handle */
        dl_memset(hdl, 0, sizeof(dl_handle_t));
    }
    
    return 0;
}

/*
 * dlerror - Get error message
 */
char *dlerror(void) {
    char *error;
    
    if (!dl_has_error) {
        return NULL;
    }
    
    error = dl_error_buf;
    dl_clear_error();
    return error;
}

/*
 * Implementation Notes:
 * 
 * 1. This implementation provides basic ELF64 shared library loading.
 * 2. Symbol tables are loaded during dlopen and cached in the handle.
 * 3. dlsym uses cached symbol tables for fast lookup.
 * 4. Reference counting allows multiple dlopen calls.
 * 5. Memory is allocated using mmap for proper permissions.
 * 
 * Limitations:
 * - Simplified relocation handling (relocations not processed)
 * - No support for lazy binding (RTLD_LAZY)
 * - No support for symbol interposition
 * - Fixed-size handle table (MAX_LOADED_LIBRARIES)
 * - Basic error messages
 */
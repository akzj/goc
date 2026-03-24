/*
 * dlfcn_fixed.c - Fixed ELF64 loader for testing
 */

#define _GNU_SOURCE
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <dlfcn.h>
#include <fcntl.h>
#include <unistd.h>
#include <sys/mman.h>
#include <stdint.h>
#include <elf.h>

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

static void dl_set_error(const char *msg) {
    strncpy(dl_error_buf, msg, MAX_ERROR_MSG - 1);
    dl_error_buf[MAX_ERROR_MSG - 1] = '\0';
    dl_has_error = 1;
}

static void dl_clear_error(void) {
    dl_error_buf[0] = '\0';
    dl_has_error = 0;
}

static int dl_find_free_handle(void) {
    int i;
    for (i = 0; i < MAX_LOADED_LIBRARIES; i++) {
        if (!handles[i].in_use) {
            return i;
        }
    }
    return -1;
}

static int dl_find_handle_by_path(const char *path) {
    int i;
    for (i = 0; i < MAX_LOADED_LIBRARIES; i++) {
        if (handles[i].in_use && handles[i].path && 
            strcmp(handles[i].path, path) == 0) {
            return i;
        }
    }
    return -1;
}

/* ============================================================================
 * Main Implementation
 * ============================================================================ */

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
    
    if (!filename) {
        dl_set_error("dlopen: filename is NULL");
        return NULL;
    }
    
    /* Check if already loaded */
    handle_idx = dl_find_handle_by_path(filename);
    if (handle_idx >= 0) {
        handles[handle_idx].refcount++;
        printf("  [dlopen] Library already loaded, refcount=%d\n", handles[handle_idx].refcount);
        return &handles[handle_idx];
    }
    
    /* Find free handle slot */
    handle_idx = dl_find_free_handle();
    if (handle_idx < 0) {
        dl_set_error("dlopen: too many libraries loaded");
        return NULL;
    }
    
    /* Open file */
    fd = open(filename, O_RDONLY);
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
    
    if (lseek(fd, ehdr.e_phoff, SEEK_SET) < 0) {
        dl_set_error("dlopen: cannot seek to program headers");
        close(fd);
        free(phdrs);
        return NULL;
    }
    
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
    
    printf("  [dlopen] min_vaddr=0x%lx, max_vaddr=0x%lx, total_size=0x%lx\n", min_vaddr, max_vaddr, total_size);
    
    /* Allocate memory for the library */
    map_base = mmap(NULL, total_size, 
                    PROT_READ | PROT_WRITE | PROT_EXEC,
                    MAP_PRIVATE | MAP_ANONYMOUS,
                    -1, 0);
    
    if (map_base == MAP_FAILED || map_base == NULL) {
        dl_set_error("dlopen: cannot allocate memory");
        close(fd);
        free(phdrs);
        return NULL;
    }
    
    printf("  [dlopen] map_base=%p\n", map_base);
    
    /* Load segments */
    for (i = 0; i < ehdr.e_phnum; i++) {
        if (phdrs[i].p_type == PT_LOAD && phdrs[i].p_filesz > 0) {
            void *seg_addr = (char *)map_base + phdrs[i].p_vaddr - min_vaddr;
            
            if (lseek(fd, phdrs[i].p_offset, SEEK_SET) < 0) {
                continue;
            }
            
            bytes_read = read(fd, seg_addr, phdrs[i].p_filesz);
            printf("  [dlopen] Loaded segment %d: offset=0x%lx, vaddr=0x%lx, filesz=0x%lx, seg_addr=%p, read=%zd\n",
                   i, phdrs[i].p_offset, phdrs[i].p_vaddr, phdrs[i].p_filesz, seg_addr, bytes_read);
        }
    }
    
    /* Read section headers to find symbol tables */
    if (ehdr.e_shoff != 0 && ehdr.e_shnum != 0) {
        shdrs = (Elf64_Shdr *)malloc(ehdr.e_shnum * sizeof(Elf64_Shdr));
        if (shdrs) {
            if (lseek(fd, ehdr.e_shoff, SEEK_SET) >= 0) {
                if (read(fd, shdrs, ehdr.e_shnum * sizeof(Elf64_Shdr)) == 
                    (ssize_t)(ehdr.e_shnum * sizeof(Elf64_Shdr))) {
                    
                    /* Find .dynsym and .dynstr sections */
                    for (i = 0; i < ehdr.e_shnum; i++) {
                        if (shdrs[i].sh_type == SHT_DYNSYM) {
                            symtab = (char *)map_base + shdrs[i].sh_addr - min_vaddr;
                            symcount = shdrs[i].sh_size / sizeof(Elf64_Sym);
                            printf("  [dlopen] Found DYNSYM: section=%d, addr=0x%lx, symtab=%p, symcount=%lu\n",
                                   i, shdrs[i].sh_addr, symtab, symcount);
                            
                            /* Use sh_link to find corresponding strtab */
                            if (shdrs[i].sh_link < ehdr.e_shnum) {
                                strtab = (char *)map_base + shdrs[shdrs[i].sh_link].sh_addr - min_vaddr;
                                printf("  [dlopen] Found STRTAB: section=%lu, addr=0x%lx, strtab=%p\n",
                                       shdrs[i].sh_link, shdrs[shdrs[i].sh_link].sh_addr, strtab);
                            }
                            break;
                        }
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
    handles[handle_idx].path = strdup(filename);
    
    printf("  [dlopen] Handle initialized: base=%p, symtab=%p, strtab=%p\n", 
           map_base, symtab, strtab);
    
    return &handles[handle_idx];
}

void *dlsym(void *handle, const char *symbol) {
    dl_handle_t *hdl;
    Elf64_Sym *symtab;
    char *strtab;
    size_t i;
    
    dl_clear_error();
    
    if (!handle) {
        dl_set_error("dlsym: invalid handle");
        return NULL;
    }
    
    hdl = (dl_handle_t *)handle;
    
    if (!hdl->in_use) {
        dl_set_error("dlsym: invalid handle (not loaded)");
        return NULL;
    }
    
    if (!hdl->symtab || !hdl->strtab) {
        dl_set_error("dlsym: no symbol table");
        return NULL;
    }
    
    symtab = (Elf64_Sym *)hdl->symtab;
    strtab = (char *)hdl->strtab;
    
    printf("  [dlsym] Searching for '%s' in %lu symbols\n", symbol, hdl->symcount);
    printf("  [dlsym] symtab=%p, strtab=%p, base=%p\n", symtab, strtab, hdl->base);
    
    /* Search for symbol */
    for (i = 0; i < hdl->symcount; i++) {
        const char *sym_name = strtab + symtab[i].st_name;
        
        printf("  [dlsym] Symbol %lu: name='%s', st_value=0x%lx, st_shndx=%d\n",
               i, sym_name, symtab[i].st_value, symtab[i].st_shndx);
        
        if (strcmp(sym_name, symbol) == 0) {
            /* Found symbol */
            if (symtab[i].st_shndx == 0) {
                printf("  [dlsym] Symbol '%s' is undefined (st_shndx=0)\n", symbol);
                continue;
            }
            
            void *sym_addr = (char *)hdl->base + symtab[i].st_value - hdl->min_vaddr;
            printf("  [dlsym] Found '%s' at %p (st_value=0x%lx, min_vaddr=0x%lx)\n",
                   symbol, sym_addr, symtab[i].st_value, hdl->min_vaddr);
            return sym_addr;
        }
    }
    
    dl_set_error("dlsym: symbol not found");
    return NULL;
}

int dlclose(void *handle) {
    dl_handle_t *hdl;
    
    dl_clear_error();
    
    if (!handle) {
        dl_set_error("dlclose: invalid handle");
        return -1;
    }
    
    hdl = (dl_handle_t *)handle;
    
    if (!hdl->in_use) {
        dl_set_error("dlclose: invalid handle (not loaded)");
        return -1;
    }
    
    hdl->refcount--;
    
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
        memset(hdl, 0, sizeof(dl_handle_t));
    }
    
    return 0;
}

char *dlerror(void) {
    char *error;
    
    if (!dl_has_error) {
        return NULL;
    }
    
    error = dl_error_buf;
    dl_clear_error();
    return error;
}

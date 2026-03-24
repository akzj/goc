/*
 * dlfcn_standalone.c - Standalone ELF64 loader for testing
 * 
 * This is a standalone version for testing with GCC.
 * Uses system headers instead of GOC syscall_wrapper.
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>
#include <sys/mman.h>
#include <stdint.h>

/* ============================================================================
 * ELF64 Structures
 * ============================================================================ */

#define EI_NIDENT 16

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

typedef struct {
    uint32_t st_name;
    unsigned char st_info;
    unsigned char st_other;
    uint16_t st_shndx;
    uint64_t st_value;
    uint64_t st_size;
} Elf64_Sym;

typedef struct {
    int64_t d_tag;
    uint64_t d_val;
} Elf64_Dyn;

/* ELF Constants */
#define ELFMAG0 0x7f
#define ELFMAG1 'E'
#define ELFMAG2 'L'
#define ELFMAG3 'F'
#define ELFCLASS64 2
#define ELFDATA2LSB 1
#define ET_DYN 3
#define EM_X86_64 62
#define PT_NULL    0
#define PT_LOAD    1
#define PT_DYNAMIC 2
#define SHT_NULL     0
#define SHT_PROGBITS 1
#define SHT_SYMTAB   2
#define SHT_STRTAB   3
#define SHT_RELA     4
#define SHT_NOBITS   8
#define SHT_REL      9
#define SHT_DYNSYM   11
#define DT_NULL     0
#define DT_SYMTAB   6
#define DT_STRTAB   5
#define ELF64_ST_BIND(info) (((unsigned char)(info)) >> 4)
#define ELF64_ST_TYPE(info) ((info) & 0xf)
#define ELF64_R_SYM(info) ((info) >> 32)
#define ELF64_R_TYPE(info) ((info) & 0xffffffff)

/* ============================================================================
 * Handle Structure
 * ============================================================================ */

#define MAX_LOADED_LIBRARIES 64
#define MAX_ERROR_MSG 256

typedef struct dl_handle {
    void *base;
    char *path;
    int refcount;
    void *symtab;
    void *strtab;
    size_t symcount;
    size_t symtab_size;
    size_t strtab_size;
    size_t min_vaddr;  /* Minimum virtual address for address translation */
    int in_use;
} dl_handle_t;

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

static void dl_debug_flush(void) {
    fflush(stdout);
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
    ssize_t bytes_read;
    int i;
    void *map_base = NULL;
    void *dyn_section = NULL;
    void *symtab = NULL;
    void *strtab = NULL;
    size_t symtab_size = 0;
    size_t strtab_size = 0;
    size_t total_size = 0;
    size_t min_vaddr = 0xFFFFFFFFFFFFFFFFULL;
    size_t max_vaddr = 0;
    
    (void)flag;  /* Simplified - ignore flags for now */
    
    dl_clear_error();
    
    if (!filename) {
        dl_set_error("dlopen: filename is NULL");
        return NULL;
    }
    
    handle_idx = dl_find_handle_by_path(filename);
    if (handle_idx >= 0) {
        handles[handle_idx].refcount++;
        return &handles[handle_idx];
    }
    
    handle_idx = dl_find_free_handle();
    if (handle_idx < 0) {
        dl_set_error("dlopen: too many libraries loaded");
        return NULL;
    }
    
    fd = open(filename, O_RDONLY);
    if (fd < 0) {
        dl_set_error("dlopen: cannot open file");
        return NULL;
    }
    
    bytes_read = read(fd, &ehdr, sizeof(ehdr));
    if (bytes_read != sizeof(ehdr)) {
        dl_set_error("dlopen: cannot read ELF header");
        close(fd);
        return NULL;
    }
    
    if (ehdr.e_ident[0] != ELFMAG0 || ehdr.e_ident[1] != ELFMAG1 ||
        ehdr.e_ident[2] != ELFMAG2 || ehdr.e_ident[3] != ELFMAG3) {
        dl_set_error("dlopen: not an ELF file");
        close(fd);
        return NULL;
    }
    
    if (ehdr.e_ident[4] != ELFCLASS64) {
        dl_set_error("dlopen: not a 64-bit ELF file");
        close(fd);
        return NULL;
    }
    
    if (ehdr.e_ident[5] != ELFDATA2LSB) {
        dl_set_error("dlopen: unsupported byte order");
        close(fd);
        return NULL;
    }
    
    if (ehdr.e_type != ET_DYN) {
        dl_set_error("dlopen: not a shared object file");
        close(fd);
        return NULL;
    }
    
    if (ehdr.e_machine != EM_X86_64) {
        dl_set_error("dlopen: not x86-64 architecture");
        close(fd);
        return NULL;
    }
    
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
    
    printf("  [DEBUG] min_vaddr=0x%lx, max_vaddr=0x%lx, total_size=0x%lx\n", min_vaddr, max_vaddr, total_size);
    
    map_base = mmap(NULL, total_size, 
                    PROT_READ | PROT_WRITE | PROT_EXEC,
                    MAP_PRIVATE | MAP_ANONYMOUS,
                    -1, 0);
    
    printf("  [DEBUG] map_base=%p\n", map_base);
    
    if (map_base == MAP_FAILED || map_base == NULL) {
        dl_set_error("dlopen: cannot allocate memory");
        close(fd);
        free(phdrs);
        return NULL;
    }
    
    for (i = 0; i < ehdr.e_phnum; i++) {
        if (phdrs[i].p_type == PT_LOAD && phdrs[i].p_filesz > 0) {
            void *seg_addr = (char *)map_base + phdrs[i].p_vaddr - min_vaddr;
            
            printf("  [DEBUG] Loading segment %d: offset=0x%lx, vaddr=0x%lx, filesz=0x%lx, seg_addr=%p\n",
                   i, phdrs[i].p_offset, phdrs[i].p_vaddr, phdrs[i].p_filesz, seg_addr);
            
            if (lseek(fd, phdrs[i].p_offset, SEEK_SET) < 0) {
                continue;
            }
            
            bytes_read = read(fd, seg_addr, phdrs[i].p_filesz);
            printf("  [DEBUG] Read %ld bytes\n", bytes_read);
        }
    }
    
    close(fd);
    
    for (i = 0; i < ehdr.e_phnum; i++) {
        if (phdrs[i].p_type == PT_DYNAMIC) {
            dyn_section = (char *)map_base + phdrs[i].p_vaddr - min_vaddr;
            break;
        }
    }
    
    if (dyn_section) {
        Elf64_Dyn *dyn = (Elf64_Dyn *)dyn_section;
        while (dyn->d_tag != DT_NULL) {
            switch (dyn->d_tag) {
                case DT_SYMTAB:
                    symtab = (void *)((char *)map_base + dyn->d_val - min_vaddr);
                    break;
                case DT_STRTAB:
                    strtab = (void *)((char *)map_base + dyn->d_val - min_vaddr);
                    break;
            }
            dyn++;
        }
    }
    
    /* Estimate symbol count from dynamic section */
    symtab_size = 0;
    strtab_size = 0;
    
    handles[handle_idx].in_use = 1;
    handles[handle_idx].base = map_base;
    handles[handle_idx].refcount = 1;
    handles[handle_idx].symtab = symtab;
    handles[handle_idx].strtab = strtab;
    handles[handle_idx].symcount = 0;
    handles[handle_idx].symtab_size = symtab_size;
    handles[handle_idx].strtab_size = strtab_size;
    handles[handle_idx].min_vaddr = min_vaddr;
    
    handles[handle_idx].path = strdup(filename);
    
    free(phdrs);
    
    return &handles[handle_idx];
}

void *dlsym(void *handle, const char *symbol) {
    dl_handle_t *hdl;
    Elf64_Sym *symtab;
    char *strtab;
    size_t i;
    
    /* Simplified symbol lookup - search section headers for .dynsym */
    int fd;
    Elf64_Ehdr ehdr;
    Elf64_Shdr *shdrs = NULL;
    void *symtab_data = NULL;
    void *strtab_data = NULL;
    size_t symcount = 0;
    
    dl_clear_error();
    
    if (!handle) {
        dl_set_error("dlsym: invalid handle");
        return NULL;
    }
    
    hdl = (dl_handle_t *)handle;
    
    if (!hdl->in_use || !hdl->path) {
        dl_set_error("dlsym: invalid handle");
        return NULL;
    }
    
    printf("  [dlsym DEBUG] Opening file: %s\n", hdl->path);
    
    /* Re-open file to read section headers */
    fd = open(hdl->path, O_RDONLY);
    if (fd < 0) {
        dl_set_error("dlsym: cannot open file");
        return NULL;
    }
    
    printf("  [dlsym DEBUG] File opened, fd=%d\n", fd);
    
    if (read(fd, &ehdr, sizeof(ehdr)) != sizeof(ehdr)) {
        close(fd);
        dl_set_error("dlsym: cannot read ELF header");
        return NULL;
    }
    
    printf("  [dlsym DEBUG] ELF header read, e_shoff=0x%lx, e_shnum=%d\n", ehdr.e_shoff, ehdr.e_shnum);
    
    if (ehdr.e_shoff == 0 || ehdr.e_shnum == 0) {
        close(fd);
        dl_set_error("dlsym: no section headers (stripped library)");
        return NULL;
    }
    
    shdrs = (Elf64_Shdr *)malloc(ehdr.e_shnum * sizeof(Elf64_Shdr));
    if (!shdrs) {
        close(fd);
        dl_set_error("dlsym: out of memory");
        return NULL;
    }
    
    printf("  [dlsym DEBUG] Allocated shdrs at %p\n", shdrs);
    
    if (lseek(fd, ehdr.e_shoff, SEEK_SET) < 0) {
        close(fd);
        free(shdrs);
        dl_set_error("dlsym: cannot seek to section headers");
        return NULL;
    }
    
    if (read(fd, shdrs, ehdr.e_shnum * sizeof(Elf64_Shdr)) != 
        (ssize_t)(ehdr.e_shnum * sizeof(Elf64_Shdr))) {
        close(fd);
        free(shdrs);
        dl_set_error("dlsym: cannot read section headers");
        return NULL;
    }
    
    printf("  [dlsym DEBUG] Section headers read\n");
    
    /* Find .dynsym and .dynstr sections */
    /* Note: sh_addr contains virtual addresses, need to adjust by base offset */
    for (i = 0; i < ehdr.e_shnum; i++) {
        printf("  [dlsym DEBUG] Section %d: type=%d, addr=0x%lx, size=0x%lx\n",
               i, shdrs[i].sh_type, shdrs[i].sh_addr, shdrs[i].sh_size);
        if (shdrs[i].sh_type == SHT_DYNSYM) {
            /* sh_addr is virtual address, convert to offset in mapped memory */
            symtab_data = (char *)hdl->base + shdrs[i].sh_addr - hdl->min_vaddr;
            symcount = shdrs[i].sh_size / sizeof(Elf64_Sym);
            printf("  [dlsym DEBUG] Found DYNSYM: symtab_data=%p, symcount=%lu\n", symtab_data, symcount);
        }
    }
    
    /* Use sh_link to find corresponding strtab */
    for (i = 0; i < ehdr.e_shnum; i++) {
        if (shdrs[i].sh_type == SHT_DYNSYM) {
            size_t strtab_idx = shdrs[i].sh_link;
            printf("  [dlsym DEBUG] DYNSYM sh_link=%lu\n", strtab_idx);
            if (strtab_idx < ehdr.e_shnum) {
                strtab_data = (char *)hdl->base + shdrs[strtab_idx].sh_addr - hdl->min_vaddr;
                printf("  [dlsym DEBUG] Found STRTAB: strtab_data=%p\n", strtab_data);
            }
            break;
        }
    }
    
    close(fd);
    free(shdrs);
    
    printf("  [dlsym DEBUG] After free: symtab_data=%p, strtab_data=%p, symcount=%lu\n", 
           symtab_data, strtab_data, symcount);
    
    if (!symtab_data || !strtab_data) {
        dl_set_error("dlsym: no symbol table");
        return NULL;
    }
    
    symtab = (Elf64_Sym *)symtab_data;
    strtab = (char *)strtab_data;
    
    printf("  [dlsym DEBUG] Searching for symbol: %s\n", symbol);
    printf("  [dlsym DEBUG] symtab[0].st_name=%u, strtab+st_name=%s\n", 
           symtab[0].st_name, strtab + symtab[0].st_name);
    
    for (i = 0; i < symcount; i++) {
        const char *sym_name = strtab + symtab[i].st_name;
        printf("  [dlsym DEBUG] Symbol %lu: name=%s, st_value=0x%lx, st_shndx=%d\n",
               i, sym_name, symtab[i].st_value, symtab[i].st_shndx);
        
        if (strcmp(sym_name, symbol) == 0) {
            if (symtab[i].st_shndx == 0) {
                printf("  [dlsym DEBUG] Skipping undefined symbol\n");
                continue;  /* Skip undefined symbols */
            }
            /* st_value is a virtual address, convert to actual address */
            printf("  [dlsym DEBUG] Found symbol! Returning address %p\n",
                   (void *)((char *)hdl->base + symtab[i].st_value - hdl->min_vaddr));
            return (void *)((char *)hdl->base + symtab[i].st_value - hdl->min_vaddr);
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
        dl_set_error("dlclose: invalid handle");
        return -1;
    }
    
    hdl->refcount--;
    
    if (hdl->refcount == 0) {
        if (hdl->base && hdl->base != MAP_FAILED) {
            munmap(hdl->base, 1024 * 1024);
        }
        
        if (hdl->path) {
            free(hdl->path);
        }
        
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
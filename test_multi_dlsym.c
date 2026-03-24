#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>
#include <sys/mman.h>
#include <stdint.h>

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

#define ELFMAG0 0x7f
#define ELFMAG1 'E'
#define ELFMAG2 'L'
#define ELFMAG3 'F'
#define ELFCLASS64 2
#define ELFDATA2LSB 1
#define ET_DYN 3
#define EM_X86_64 62
#define PT_LOAD    1
#define SHT_DYNSYM   11

typedef struct {
    void *base;
    size_t min_vaddr;
} handle_t;

void *do_dlopen(const char *filename) {
    int fd = open(filename, O_RDONLY);
    Elf64_Ehdr ehdr;
    Elf64_Phdr *phdrs;
    void *map_base;
    size_t min_vaddr = 0xFFFFFFFFFFFFFFFFULL;
    size_t max_vaddr = 0;
    int i;
    
    read(fd, &ehdr, sizeof(ehdr));
    
    phdrs = malloc(ehdr.e_phnum * sizeof(Elf64_Phdr));
    lseek(fd, ehdr.e_phoff, SEEK_SET);
    read(fd, phdrs, ehdr.e_phnum * sizeof(Elf64_Phdr));
    
    for (i = 0; i < ehdr.e_phnum; i++) {
        if (phdrs[i].p_type == PT_LOAD) {
            if (phdrs[i].p_vaddr < min_vaddr) min_vaddr = phdrs[i].p_vaddr;
            if (phdrs[i].p_vaddr + phdrs[i].p_memsz > max_vaddr) max_vaddr = phdrs[i].p_vaddr + phdrs[i].p_memsz;
        }
    }
    
    map_base = mmap(NULL, max_vaddr - min_vaddr, PROT_READ | PROT_WRITE | PROT_EXEC, MAP_PRIVATE | MAP_ANONYMOUS, -1, 0);
    
    for (i = 0; i < ehdr.e_phnum; i++) {
        if (phdrs[i].p_type == PT_LOAD && phdrs[i].p_filesz > 0) {
            void *seg_addr = (char *)map_base + phdrs[i].p_vaddr - min_vaddr;
            lseek(fd, phdrs[i].p_offset, SEEK_SET);
            read(fd, seg_addr, phdrs[i].p_filesz);
        }
    }
    
    close(fd);
    free(phdrs);
    
    handle_t *h = malloc(sizeof(handle_t));
    h->base = map_base;
    h->min_vaddr = min_vaddr;
    return h;
}

void *do_dlsym(void *handle, const char *symbol) {
    handle_t *h = (handle_t *)handle;
    int fd;
    Elf64_Ehdr ehdr;
    Elf64_Shdr *shdrs;
    int i;
    
    printf("  [dlsym] Opening file\n"); fflush(stdout);
    fd = open("test_dlfcn/test_lib.so", O_RDONLY);
    printf("  [dlsym] Reading ELF header\n"); fflush(stdout);
    read(fd, &ehdr, sizeof(ehdr));
    
    printf("  [dlsym] Reading section headers (e_shnum=%d)\n", ehdr.e_shnum); fflush(stdout);
    shdrs = malloc(ehdr.e_shnum * sizeof(Elf64_Shdr));
    lseek(fd, ehdr.e_shoff, SEEK_SET);
    read(fd, shdrs, ehdr.e_shnum * sizeof(Elf64_Shdr));
    
    printf("  [dlsym] Searching for DYNSYM\n"); fflush(stdout);
    for (i = 0; i < ehdr.e_shnum; i++) {
        if (shdrs[i].sh_type == SHT_DYNSYM) {
            printf("  [dlsym] Found DYNSYM, building symtab/strtab\n"); fflush(stdout);
            Elf64_Sym *symtab = (Elf64_Sym *)((char *)h->base + shdrs[i].sh_addr - h->min_vaddr);
            char *strtab = (char *)h->base + shdrs[shdrs[i].sh_link].sh_addr - h->min_vaddr;
            size_t symcount = shdrs[i].sh_size / sizeof(Elf64_Sym);
            size_t j;
            
            printf("  [dlsym] symtab=%p, strtab=%p, symcount=%lu, searching for '%s'\n", 
                   symtab, strtab, symcount, symbol); fflush(stdout);
            
            for (j = 0; j < symcount; j++) {
                const char *name = strtab + symtab[j].st_name;
                if (strcmp(name, symbol) == 0) {
                    printf("  [dlsym] Found symbol '%s' at st_value=0x%lx\n", symbol, symtab[j].st_value); fflush(stdout);
                    if (symtab[j].st_shndx == 0) {
                        printf("  [dlsym] Skipping undefined symbol\n"); fflush(stdout);
                        close(fd);
                        free(shdrs);
                        return NULL;
                    }
                    void *addr = (char *)h->base + symtab[j].st_value - h->min_vaddr;
                    printf("  [dlsym] Returning addr=%p\n", addr); fflush(stdout);
                    close(fd);
                    free(shdrs);
                    return addr;
                }
            }
            break;
        }
    }
    
    printf("  [dlsym] Symbol not found\n"); fflush(stdout);
    close(fd);
    free(shdrs);
    return NULL;
}

int main() {
    void *h;
    void *sym;
    
    printf("Test 1: dlopen\n");
    h = do_dlopen("test_dlfcn/test_lib.so");
    printf("  handle=%p, base=%p, min_vaddr=0x%lx\n\n", h, ((handle_t*)h)->base, ((handle_t*)h)->min_vaddr);
    
    printf("Test 2: dlsym test_add\n");
    sym = do_dlsym(h, "test_add");
    printf("  result=%p\n", sym);
    if (sym) {
        typedef int (*add_fn)(int, int);
        printf("  test_add(5,3)=%d\n\n", ((add_fn)sym)(5, 3));
    }
    
    printf("Test 3: dlsym test_multiply\n");
    sym = do_dlsym(h, "test_multiply");
    printf("  result=%p\n", sym);
    if (sym) {
        typedef int (*mul_fn)(int, int);
        printf("  test_multiply(4,7)=%d\n\n", ((mul_fn)sym)(4, 7));
    }
    
    printf("Test 4: dlsym test_message\n");
    sym = do_dlsym(h, "test_message");
    printf("  result=%p\n", sym);
    if (sym) {
        printf("  *test_message=%s\n\n", *(const char **)sym);
    }
    
    printf("Done!\n");
    return 0;
}

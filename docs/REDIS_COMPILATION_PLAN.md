# Redis Compilation with GOC - Development Plan

## Executive Summary

This document outlines a comprehensive multi-phase development plan to enable the GOC C11 compiler to compile the Redis source code. Redis is a complex, production-grade C project that uses advanced C features, threading, system calls, and external dependencies.

**Current State**: GOC is a functional C11 compiler with basic pipeline (lexer → parser → semantic analyzer → IR → codegen → linker → ELF64).

**Target State**: GOC can compile Redis core (redis-server, redis-cli) with full functionality.

---

## Phase 0: Redis Source Analysis (COMPLETED)

### 0.1 Source Structure
- **Location**: `/home/ubuntu/workspace/goc/redis-source/`
- **Core files**: ~100 C files in `src/` directory
- **Dependencies**: `deps/` (jemalloc, lua, hiredis, linenoise, hdr_histogram, fpconv, xxhash)
- **Main targets**: redis-server, redis-cli, redis-check-rdb, redis-check-aof, redis-benchmark

### 0.2 Key C Features Used by Redis

| Feature | Usage Level | Files | Priority |
|---------|-------------|-------|----------|
| C11 Atomics | Heavy | atomicvar.h, bio.c, server.c | CRITICAL |
| pthreads | Heavy | bio.c, iothread.c, threads_mngr.c | CRITICAL |
| System calls | Heavy | ae.c, anet.c, socket.c, unix.c | CRITICAL |
| Complex structs | Heavy | sds.c, dict.c, rax.c, quicklist.c | CRITICAL |
| Function pointers | Heavy | module.c, commands.c, networking.c | HIGH |
| Variable args | Medium | server.c, util.c | HIGH |
| Unions | Medium | object.c, networking.c | MEDIUM |
| Bitfields | Medium | Various | MEDIUM |
| Inline assembly | Light | cluster_asm.c | LOW |
| TLS/SSL | Optional | tls.c | OPTIONAL |
| Lua integration | Heavy | script_lua.c, function_lua.c | HIGH |

### 0.3 External Dependencies Analysis

| Dependency | Purpose | Complexity | GOC Action |
|------------|---------|------------|------------|
| jemalloc | Memory allocator | Very High | Replace with libc or implement subset |
| lua | Scripting engine | High | Compile with GOC or provide stubs |
| hiredis | Redis protocol lib | Medium | Compile with GOC |
| linenoise | CLI editing | Low | Compile with GOC |
| hdr_histogram | Latency metrics | Medium | Compile with GOC |
| fpconv | Float conversion | Low | Compile with GOC |
| xxhash | Hashing | Low | Compile with GOC |

---

## Phase 1: GOC Core Extensions (Estimated: 4-6 weeks)

### 1.1 C11 Atomics Support
**Goal**: Implement full C11 `_Atomic` and `stdatomic.h` support

**Tasks**:
- [ ] Add atomic type handling in lexer/parser
- [ ] Implement atomic operations in semantic analyzer
- [ ] Generate lock-free x86-64 atomic instructions (LOCK prefix, CMPXCHG, XADD, etc.)
- [ ] Implement memory ordering semantics (acquire, release, seq_cst)
- [ ] Add stdatomic.h to GOC standard library

**Acceptance Criteria**:
- Can compile atomicvar.h without errors
- Atomic increment/decrement produces correct LOCK-prefixed instructions
- Atomic compare-and-swap works correctly
- All Redis atomic macros compile and link

**Files to modify**:
- `pkg/parser/expr.y` - atomic expressions
- `pkg/semantic/analyzer.go` - atomic type checking
- `pkg/codegen/x86.go` - atomic instruction emission
- `pkg/stdlib/stdatomic.h` - new standard header

### 1.2 pthreads Support
**Goal**: Implement pthread library compatibility layer

**Tasks**:
- [ ] Create pthread.h header with all Redis-used functions
- [ ] Implement pthread_create, pthread_join, pthread_mutex_*, pthread_cond_*
- [ ] Map pthread primitives to Go runtime or direct syscalls
- [ ] Handle thread-local storage (TLS)
- [ ] Implement signal handling for threads

**Acceptance Criteria**:
- bio.c compiles and links
- Thread creation and synchronization works
- No race conditions in basic tests
- Redis background I/O threads function correctly

**Files to modify**:
- `pkg/stdlib/pthread.h` - new standard header
- `pkg/codegen/runtime.c` - pthread runtime support
- `pkg/linker/symbols.go` - pthread symbol resolution

### 1.3 System Call Interface
**Goal**: Complete Linux x86-64 syscall support for Redis

**Tasks**:
- [ ] Implement syscall wrapper generation
- [ ] Add all Redis-used syscalls: socket, bind, listen, accept, epoll_*, read, write, etc.
- [ ] Handle errno and error codes correctly
- [ ] Implement signal handling (signal, sigaction, sigprocmask)
- [ ] Support file descriptor operations (fcntl, dup, close)

**Acceptance Criteria**:
- ae.c (event loop) compiles and functions
- anet.c (networking) compiles and functions
- Redis can accept network connections
- Event-driven I/O works correctly

**Files to modify**:
- `pkg/stdlib/unistd.h`, `pkg/stdlib/sys/socket.h`, etc.
- `pkg/codegen/syscall.go` - syscall instruction emission
- `pkg/linker/elf.go` - syscall stub linking

---

## Phase 2: Data Structure Compatibility (Estimated: 3-4 weeks)

### 2.1 SDS (Simple Dynamic Strings)
**Goal**: Full SDS library compatibility

**Tasks**:
- [ ] Verify sds.c compiles without modifications
- [ ] Test all SDS operations (append, prepend, trim, etc.)
- [ ] Optimize common SDS patterns
- [ ] Ensure memory safety with GOC's allocator

**Acceptance Criteria**:
- sds.c compiles with 100% test pass rate
- All SDS macros and functions work correctly
- No memory leaks or corruption

### 2.2 Dict (Hash Tables)
**Goal**: Full dict library compatibility

**Tasks**:
- [ ] Verify dict.c compiles
- [ ] Test hash table operations (add, find, delete, resize)
- [ ] Verify incremental rehashing works
- [ ] Test with Redis workloads

**Acceptance Criteria**:
- dict.c compiles and passes all tests
- Hash table performance within 20% of GCC compilation

### 2.3 RAX (Radix Tree)
**Goal**: Full rax library compatibility

**Tasks**:
- [ ] Verify rax.c compiles
- [ ] Test radix tree operations
- [ ] Verify iterator functionality
- [ ] Test with complex keys

**Acceptance Criteria**:
- rax.c compiles and functions correctly
- All rax API calls work as expected

### 2.4 Quicklist & Ziplist
**Goal**: Complete list data structure support

**Tasks**:
- [ ] Verify quicklist.c, ziplist.c, listpack.c compile
- [ ] Test compression and encoding
- [ ] Verify memory efficiency
- [ ] Test with Redis list operations

**Acceptance Criteria**:
- All list data structures compile and function
- Redis list commands work correctly

---

## Phase 3: Build System & Dependencies (Estimated: 3-4 weeks)

### 3.1 Makefile Integration
**Goal**: GOC can build Redis using modified Makefile

**Tasks**:
- [ ] Create `Makefile.goc` based on Redis Makefile
- [ ] Replace GCC with GOC in build commands
- [ ] Handle dependency compilation order
- [ ] Integrate with GOC's linker

**Acceptance Criteria**:
- `make -f Makefile.goc` builds redis-server
- All object files link correctly
- Build completes without manual intervention

### 3.2 Dependency Compilation
**Goal**: Compile all Redis dependencies with GOC

**Tasks**:
- [ ] Compile linenoise (CLI editing)
- [ ] Compile fpconv (float conversion)
- [ ] Compile xxhash (hashing)
- [ ] Compile hdr_histogram (metrics)
- [ ] Compile hiredis (protocol library)
- [ ] Handle jemalloc (replace with libc or minimal implementation)
- [ ] Handle Lua (compile or provide stubs)

**Acceptance Criteria**:
- All dependencies compile with GOC
- Dependencies link correctly with Redis core
- No external GCC dependencies remain

### 3.3 Lua Integration (Optional Phase 3b)
**Goal**: Support Lua scripting in Redis

**Tasks**:
- [ ] Compile Lua 5.1 with GOC
- [ ] Verify Lua C API compatibility
- [ ] Test script_lua.c integration
- [ ] Test function_lua.c integration

**Acceptance Criteria**:
- Lua scripts execute in Redis
- EVAL command works correctly
- No crashes with complex scripts

---

## Phase 4: Redis Core Compilation (Estimated: 4-6 weeks)

### 4.1 Core Server Files
**Goal**: Compile all core Redis server files

**Priority Order**:
1. **Foundation**: server.h, server.c, config.h, config.c
2. **Networking**: networking.c, anet.c, ae.c, socket.c, connection.c
3. **Data**: db.c, object.c, dict.c, rax.c, quicklist.c, ziplist.c, listpack.c
4. **Commands**: t_string.c, t_list.c, t_set.c, t_zset.c, t_hash.c, t_stream.c
5. **Persistence**: rdb.c, aof.c
6. **Advanced**: cluster.c, replication.c, sentinel.c, acl.c

**Tasks**:
- [ ] Fix compilation errors file by file
- [ ] Handle any GOC limitations discovered
- [ ] Optimize code generation for performance
- [ ] Verify semantic correctness

**Acceptance Criteria**:
- All .c files compile to .o files
- No unresolved symbols
- All type checks pass

### 4.2 Linking & Executable Generation
**Goal**: Create working redis-server executable

**Tasks**:
- [ ] Link all object files
- [ ] Resolve all symbols
- [ ] Generate valid ELF64 executable
- [ ] Verify executable runs

**Acceptance Criteria**:
- redis-server binary is valid ELF64
- Binary starts without immediate crash
- Can respond to basic PING command

---

## Phase 5: Testing & Validation (Estimated: 4-6 weeks)

### 5.1 Unit Testing
**Goal**: Verify individual components

**Tasks**:
- [ ] Create unit tests for each Redis module
- [ ] Compare GOC output with GCC output
- [ ] Test edge cases and error handling
- [ ] Verify memory safety

**Acceptance Criteria**:
- 90%+ of unit tests pass
- No memory corruption detected
- Error handling matches GCC behavior

### 5.2 Integration Testing
**Goal**: Verify Redis functionality end-to-end

**Tasks**:
- [ ] Run Redis test suite (runtest)
- [ ] Test all Redis commands
- [ ] Test persistence (RDB/AOF)
- [ ] Test replication
- [ ] Test clustering (if implemented)

**Acceptance Criteria**:
- 80%+ of Redis tests pass
- Core functionality works (get/set, lists, hashes, etc.)
- Persistence saves and restores data correctly

### 5.3 Performance Benchmarking
**Goal**: Ensure acceptable performance

**Tasks**:
- [ ] Run redis-benchmark against GOC-built Redis
- [ ] Compare with GCC-built Redis
- [ ] Identify and fix performance bottlenecks
- [ ] Optimize hot paths

**Acceptance Criteria**:
- Performance within 30% of GCC-built Redis
- No major regressions in latency
- Memory usage within 20% of GCC build

### 5.4 Stress Testing
**Goal**: Verify stability under load

**Tasks**:
- [ ] Run extended workload tests
- [ ] Test with large datasets
- [ ] Test concurrent connections
- [ ] Test failure recovery

**Acceptance Criteria**:
- No crashes after 24+ hours of testing
- Handles 1000+ concurrent connections
- Recovers gracefully from errors

---

## Phase 6: CLI & Tools (Estimated: 2-3 weeks)

### 6.1 Redis CLI
**Goal**: Compile redis-cli with GOC

**Tasks**:
- [ ] Compile redis-cli.c
- [ ] Link with hiredis dependency
- [ ] Test all CLI commands
- [ ] Verify interactive mode

**Acceptance Criteria**:
- redis-cli connects to GOC-built redis-server
- All commands work correctly
- Interactive mode functions

### 6.2 Utility Tools
**Goal**: Compile Redis utility tools

**Tasks**:
- [ ] Compile redis-check-rdb
- [ ] Compile redis-check-aof
- [ ] Compile redis-benchmark
- [ ] Verify functionality

**Acceptance Criteria**:
- All utility tools compile and run
- Tools can analyze GOC-built Redis data files

---

## Phase 7: Documentation & Polish (Estimated: 2 weeks)

### 7.1 Documentation
**Goal**: Complete documentation for Redis compilation

**Tasks**:
- [ ] Write compilation guide
- [ ] Document GOC extensions for Redis
- [ ] Create troubleshooting guide
- [ ] Document known limitations

**Acceptance Criteria**:
- Users can follow guide to compile Redis
- Common issues are documented with solutions

### 7.2 Optimization
**Goal**: Improve compilation and runtime performance

**Tasks**:
- [ ] Profile GOC compilation of Redis
- [ ] Optimize slow compilation phases
- [ ] Add Redis-specific optimizations
- [ ] Improve error messages for Redis code

**Acceptance Criteria**:
- Redis compiles in under 5 minutes
- Error messages are actionable
- Generated code is efficient

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| C11 atomics too complex | Medium | High | Start with simple atomics, iterate |
| pthread implementation difficult | High | High | Use Go runtime for threading, map pthread API |
| jemalloc too complex | High | Medium | Use libc malloc initially, optimize later |
| Lua integration issues | Medium | Medium | Make Lua optional, provide stubs |
| Performance unacceptable | Medium | High | Profile early, optimize hot paths |
| Build system incompatibility | Low | Medium | Create custom Makefile.goc from scratch |

---

## Success Criteria

### Minimum Viable Product (MVP)
- [ ] redis-server compiles and runs
- [ ] Basic commands work (PING, SET, GET, DEL)
- [ ] Single-threaded mode functions
- [ ] No crashes on simple workloads

### Full Success
- [ ] All Redis commands work
- [ ] Persistence (RDB/AOF) functions
- [ ] Replication works
- [ ] Performance within 30% of GCC
- [ ] 80%+ of Redis test suite passes
- [ ] redis-cli and utilities compile

---

## Timeline Summary

| Phase | Duration | Dependencies |
|-------|----------|--------------|
| Phase 0: Analysis | ✅ Complete | None |
| Phase 1: Core Extensions | 4-6 weeks | None |
| Phase 2: Data Structures | 3-4 weeks | Phase 1 |
| Phase 3: Build System | 3-4 weeks | Phase 1, 2 |
| Phase 4: Core Compilation | 4-6 weeks | Phase 1-3 |
| Phase 5: Testing | 4-6 weeks | Phase 4 |
| Phase 6: CLI & Tools | 2-3 weeks | Phase 4 |
| Phase 7: Documentation | 2 weeks | Phase 5-6 |
| **Total** | **22-31 weeks** | Sequential with some parallelism |

---

## Immediate Next Steps

1. **Create trunk mission** with this development plan
2. **Prioritize Phase 1** (C11 atomics and pthreads)
3. **Set up test framework** for incremental validation
4. **Create branch missions** for each Phase 1 subtask
5. **Begin implementation** of atomic operations

---

## Appendix A: Redis Files Requiring Special Attention

### Files Using C11 Atomics
- atomicvar.h (header - defines atomic macros)
- bio.c (background I/O)
- iothread.c (I/O threads)
- threads_mngr.c (thread management)
- server.c (various atomic counters)

### Files Using pthreads
- bio.c (pthread_create, mutex, cond)
- iothread.c (pthread_create, mutex)
- threads_mngr.c (pthread_create, attr)
- tls.c (pthread if TLS enabled)

### Files Using Complex C Features
- module.c (function pointers, dynamic loading)
- script_lua.c (Lua C API, complex structs)
- cluster_asm.c (inline assembly - may need rewrite)
- networking.c (complex I/O, function pointers)

---

## Appendix B: GOC Standard Library Extensions Needed

### New Headers
- `stdatomic.h` - C11 atomics
- `pthread.h` - POSIX threads
- `sys/socket.h` - Socket operations
- `netinet/in.h` - Internet addresses
- `sys/epoll.h` - epoll (Linux)
- `sys/wait.h` - Process wait
- `sys/resource.h` - Resource limits
- `sys/un.h` - Unix domain sockets
- `systemd/sd-daemon.h` - systemd (stub)

### Runtime Support
- Atomic instruction emission
- Thread creation/management
- Signal handling
- Syscall wrappers
- errno handling

---

*Document Version: 1.0*
*Created: Based on Redis source analysis*
*Target: GOC Compiler v2.4.0+*
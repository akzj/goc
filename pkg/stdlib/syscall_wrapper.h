#ifndef _SYSCALL_WRAPPER_H
#define _SYSCALL_WRAPPER_H

#include <sys/syscall.h>
#include <sys/socket.h>
#include <sys/epoll.h>
#include <unistd.h>
#include <fcntl.h>
#include <signal.h>
#include <errno.h>

// Thread-local errno storage
__thread int __goc_errno = 0;

// Get errno location (thread-local)
static inline int* __goc_errno_location(void) {
    return &__goc_errno;
}

// Define errno as thread-local variable
#define errno (*__goc_errno_location())

// Syscall macro for x86-64
// Supports up to 6 arguments following x86-64 syscall convention:
// - syscall number: rax
// - arg1: rdi, arg2: rsi, arg3: rdx
// - arg4: r10, arg5: r8, arg6: r9
// - return: rax (negative = error)

#define __goc_syscall6(nr, arg1, arg2, arg3, arg4, arg5, arg6) \
    __goc_syscall_inline(nr, arg1, arg2, arg3, arg4, arg5, arg6)

#define __goc_syscall5(nr, arg1, arg2, arg3, arg4, arg5) \
    __goc_syscall_inline(nr, arg1, arg2, arg3, arg4, arg5, 0)

#define __goc_syscall4(nr, arg1, arg2, arg3, arg4) \
    __goc_syscall_inline(nr, arg1, arg2, arg3, arg4, 0, 0)

#define __goc_syscall3(nr, arg1, arg2, arg3) \
    __goc_syscall_inline(nr, arg1, arg2, arg3, 0, 0, 0)

#define __goc_syscall2(nr, arg1, arg2) \
    __goc_syscall_inline(nr, arg1, arg2, 0, 0, 0, 0)

#define __goc_syscall1(nr, arg1) \
    __goc_syscall_inline(nr, arg1, 0, 0, 0, 0, 0)

#define __goc_syscall0(nr) \
    __goc_syscall_inline(nr, 0, 0, 0, 0, 0, 0)

// Inline syscall implementation using direct syscall instruction
static inline long __goc_syscall_inline(long nr, long arg1, long arg2,
                                         long arg3, long arg4, long arg5,
                                         long arg6) {
    long ret;
    asm volatile (
        "movq %1, %%rax\n\t"
        "movq %2, %%rdi\n\t"
        "movq %3, %%rsi\n\t"
        "movq %4, %%rdx\n\t"
        "movq %5, %%r10\n\t"
        "movq %6, %%r8\n\t"
        "movq %7, %%r9\n\t"
        "syscall\n\t"
        "movq %%rax, %0\n\t"
        : "=m"(ret)
        : "r"(nr), "r"(arg1), "r"(arg2), "r"(arg3), "r"(arg4), "r"(arg5), "r"(arg6)
        : "rax", "rdi", "rsi", "rdx", "r10", "r8", "r9", "rcx", "r11", "memory"
    );
    return ret;
}

// Basic file I/O syscalls
static inline ssize_t read(int fd, void* buf, size_t count) {
    return __goc_syscall3(SYS_read, fd, (long)buf, count);
}

static inline ssize_t write(int fd, const void* buf, size_t count) {
    return __goc_syscall3(SYS_write, fd, (long)buf, count);
}

static inline int close(int fd) {
    return __goc_syscall1(SYS_close, fd);
}

static inline int open(const char* pathname, int flags, ...) {
    // open has variable arguments - mode is only used with O_CREAT
    // For simplicity, we use syscall3 and caller should pass mode as 3rd arg if needed
    return __goc_syscall3(SYS_open, (long)pathname, flags, 0);
}

static inline int openat(int dirfd, const char* pathname, int flags, ...) {
    return __goc_syscall4(SYS_openat, dirfd, (long)pathname, flags, 0);
}

// File control syscalls
static inline int fcntl(int fd, int cmd, ...) {
    long arg = 0;
    // fcntl has variable arguments depending on cmd
    // For simplicity, we use syscall3 and caller should pass arg as 3rd param if needed
    return __goc_syscall3(SYS_fcntl, fd, cmd, arg);
}

static inline int dup(int fd) {
    return __goc_syscall1(SYS_dup, fd);
}

static inline int dup2(int oldfd, int newfd) {
    return __goc_syscall2(SYS_dup2, oldfd, newfd);
}

// File status and manipulation
static inline int fsync(int fd) {
    return __goc_syscall1(SYS_fsync, fd);
}

static inline int fdatasync(int fd) {
    return __goc_syscall1(SYS_fdatasync, fd);
}

static inline int flock(int fd, int operation) {
    return __goc_syscall2(SYS_flock, fd, operation);
}

// File metadata
static inline int stat(const char* pathname, void* statbuf) {
    return __goc_syscall2(SYS_stat, (long)pathname, (long)statbuf);
}

static inline int fstat(int fd, void* statbuf) {
    return __goc_syscall2(SYS_fstat, fd, (long)statbuf);
}

static inline int lstat(const char* pathname, void* statbuf) {
    return __goc_syscall2(SYS_lstat, (long)pathname, (long)statbuf);
}

// File system operations
static inline int unlink(const char* pathname) {
    return __goc_syscall1(SYS_unlink, (long)pathname);
}

static inline int rmdir(const char* pathname) {
    return __goc_syscall1(SYS_rmdir, (long)pathname);
}

static inline int rename(const char* oldpath, const char* newpath) {
    return __goc_syscall2(SYS_rename, (long)oldpath, (long)newpath);
}

static inline int mkdir(const char* pathname, mode_t mode) {
    return __goc_syscall2(SYS_mkdir, (long)pathname, mode);
}

static inline int chdir(const char* path) {
    return __goc_syscall1(SYS_chdir, (long)path);
}

static inline char* getcwd(char* buf, size_t size) {
    return (char*)__goc_syscall2(SYS_getcwd, (long)buf, size);
}

// Process control
static inline pid_t fork(void) {
    return __goc_syscall0(SYS_fork);
}

static inline pid_t getpid(void) {
    return __goc_syscall0(SYS_getpid);
}

static inline pid_t getppid(void) {
    return __goc_syscall0(SYS_getppid);
}

static inline int kill(pid_t pid, int sig) {
    return __goc_syscall2(SYS_kill, pid, sig);
}

static inline pid_t wait(int* wstatus) {
    return __goc_syscall1(SYS_wait4, (long)wstatus);
}

static inline int execve(const char* filename, char* const argv[], char* const envp[]) {
    return __goc_syscall3(SYS_execve, (long)filename, (long)argv, (long)envp);
}

// Memory management
static inline void* mmap(void* addr, size_t length, int prot, int flags,
                         int fd, off_t offset) {
    return (void*)__goc_syscall6(SYS_mmap, (long)addr, length, prot, flags, fd, offset);
}

static inline int munmap(void* addr, size_t length) {
    return __goc_syscall2(SYS_munmap, (long)addr, length);
}

static inline int mprotect(void* addr, size_t len, int prot) {
    return __goc_syscall3(SYS_mprotect, (long)addr, len, prot);
}

// Socket syscalls
static inline int socket(int domain, int type, int protocol) {
    return __goc_syscall3(SYS_socket, domain, type, protocol);
}

static inline int bind(int sockfd, struct sockaddr* addr, socklen_t addrlen) {
    return __goc_syscall3(SYS_bind, sockfd, (long)addr, addrlen);
}

static inline int listen(int sockfd, int backlog) {
    return __goc_syscall2(SYS_listen, sockfd, backlog);
}

static inline int accept(int sockfd, struct sockaddr* addr, socklen_t* addrlen) {
    return __goc_syscall3(SYS_accept, sockfd, (long)addr, (long)addrlen);
}

static inline int connect(int sockfd, struct sockaddr* addr, socklen_t addrlen) {
    return __goc_syscall3(SYS_connect, sockfd, (long)addr, addrlen);
}

static inline ssize_t recv(int sockfd, void* buf, size_t len, int flags) {
    return __goc_syscall4(SYS_recv, sockfd, (long)buf, len, flags);
}

static inline ssize_t send(int sockfd, const void* buf, size_t len, int flags) {
    return __goc_syscall4(SYS_send, sockfd, (long)buf, len, flags);
}

static inline int shutdown(int sockfd, int how) {
    return __goc_syscall2(SYS_shutdown, sockfd, how);
}

static inline int getsockname(int sockfd, struct sockaddr* addr, socklen_t* addrlen) {
    return __goc_syscall3(SYS_getsockname, sockfd, (long)addr, (long)addrlen);
}

static inline int getpeername(int sockfd, struct sockaddr* addr, socklen_t* addrlen) {
    return __goc_syscall3(SYS_getpeername, sockfd, (long)addr, (long)addrlen);
}

static inline int setsockopt(int sockfd, int level, int optname,
                             const void* optval, socklen_t optlen) {
    return __goc_syscall5(SYS_setsockopt, sockfd, level, optname, (long)optval, optlen);
}

static inline int getsockopt(int sockfd, int level, int optname,
                             void* optval, socklen_t* optlen) {
    return __goc_syscall5(SYS_getsockopt, sockfd, level, optname, (long)optval, (long)optlen);
}

// epoll syscalls
static inline int epoll_create(int size) {
    return __goc_syscall1(SYS_epoll_create, size);
}

static inline int epoll_create1(int flags) {
    return __goc_syscall1(SYS_epoll_create1, flags);
}

static inline int epoll_ctl(int epfd, int op, int fd, struct epoll_event* event) {
    return __goc_syscall4(SYS_epoll_ctl, epfd, op, fd, (long)event);
}

static inline int epoll_wait(int epfd, struct epoll_event* events,
                             int maxevents, int timeout) {
    return __goc_syscall4(SYS_epoll_wait, epfd, (long)events, maxevents, timeout);
}

// Pipe syscalls
static inline int pipe(int pipefd[2]) {
    return __goc_syscall1(SYS_pipe, (long)pipefd);
}

static inline int pipe2(int pipefd[2], int flags) {
    return __goc_syscall2(SYS_pipe2, (long)pipefd, flags);
}

// Select/poll syscalls
static inline int select(int nfds, void* readfds, void* writefds,
                         void* exceptfds, struct timeval* timeout) {
    return __goc_syscall5(SYS_select, nfds, (long)readfds, (long)writefds,
                          (long)exceptfds, (long)timeout);
}

static inline int poll(struct pollfd* fds, nfds_t nfds, int timeout) {
    return __goc_syscall3(SYS_poll, (long)fds, nfds, timeout);
}

// Time syscalls
static inline int nanosleep(const struct timespec* req, struct timespec* rem) {
    return __goc_syscall2(SYS_nanosleep, (long)req, (long)rem);
}

static inline int gettimeofday(struct timeval* tv, struct timezone* tz) {
    return __goc_syscall2(SYS_gettimeofday, (long)tv, (long)tz);
}

static inline int clock_gettime(clockid_t clk_id, struct timespec* tp) {
    return __goc_syscall2(SYS_clock_gettime, clk_id, (long)tp);
}

// Signal syscalls
static inline int sigaction(int signum, const struct sigaction* act,
                            struct sigaction* oldact) {
    return __goc_syscall3(SYS_rt_sigaction, signum, (long)act, (long)oldact);
}

static inline int sigprocmask(int how, const sigset_t* set, sigset_t* oldset) {
    return __goc_syscall4(SYS_rt_sigprocmask, how, (long)set, (long)oldset, 8);
}

static inline int sigpending(sigset_t* set) {
    return __goc_syscall2(SYS_rt_sigpending, (long)set, 8);
}

// Miscellaneous
static inline long syscall(long number, ...) {
    // Generic syscall function for arbitrary syscalls
    // This is a simplified version - full implementation would need varargs handling
    return 0;
}

static inline int access(const char* pathname, int mode) {
    return __goc_syscall2(SYS_access, (long)pathname, mode);
}

static inline int chmod(const char* pathname, mode_t mode) {
    return __goc_syscall2(SYS_chmod, (long)pathname, mode);
}

static inline int chown(const char* pathname, uid_t owner, gid_t group) {
    return __goc_syscall3(SYS_chown, (long)pathname, owner, group);
}

static inline int readlink(const char* pathname, char* buf, size_t bufsiz) {
    return __goc_syscall3(SYS_readlink, (long)pathname, (long)buf, bufsiz);
}

static inline int symlink(const char* target, const char* linkpath) {
    return __goc_syscall2(SYS_symlink, (long)target, (long)linkpath);
}

#endif // _SYSCALL_WRAPPER_H
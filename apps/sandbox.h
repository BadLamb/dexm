#include <seccomp.h>  /* seccomp */
#include <unistd.h>   /* fork, exec */
#include <sys/mman.h> /* mmap */
#include <sys/prctl.h>/* prctl */

//TODO: Enable CGroups for usage limitations

void seccomp_filtered_proc() {
    // ensure none of our children will ever be granted more priv (via setuid, capabilities, ...)
    prctl(PR_SET_NO_NEW_PRIVS, 1);
    // ensure no escape is possible via ptrace
    prctl(PR_SET_DUMPABLE, 0);

    // Init the seccomp filter
    scmp_filter_ctx ctx;
    ctx = seccomp_init(SCMP_ACT_KILL);

    // Allow mmap, exit, sigreturn syscalls
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(rt_sigreturn), 0);
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(exit), 0);
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(mmap), 0);

    // Load the seccomp fliter
    seccomp_load(ctx);
}

void* start_app(char* filepath, size_t shared_mem_size, int* child_pid) {
    void* shared_mem = mmap(NULL, shared_mem_size, PROT_READ | PROT_WRITE, MAP_ANONYMOUS | MAP_SHARED, 0, 0);
    // Check if allocation was successful
    if(shared_mem == -1) {
        return 0;
    }

    int pid = fork();

    // Child process
    if(pid == 0){
        seccomp_filtered_proc(); // Start sandbox
    }

    (*child_pid) = pid;
    return shared_mem;
}
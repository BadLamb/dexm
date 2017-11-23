#include <seccomp.h>  /* seccomp */
#include <unistd.h>   /* fork, exec */
#include <sys/mman.h> /* mmap */
#include <sys/prctl.h>/* prctl */
#include <mqueue.h>   /* mq_open etc*/
#include <fcntl.h>    /* O_* constants */
//TODO: Enable CGroups for usage limitations

void seccomp_filtered_proc() {
    // ensure none of our children will ever be granted more priv (via setuid, capabilities, ...)
    prctl(PR_SET_NO_NEW_PRIVS, 1);
    // ensure no escape is possible via ptrace
    prctl(PR_SET_DUMPABLE, 0);

    // Init the seccomp filter
    scmp_filter_ctx ctx;
    ctx = seccomp_init(SCMP_ACT_KILL);

    /* Allow as few syscalls as we can. mq_* is needed for IPC with
     * the Golang process
     */
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(rt_sigreturn), 0);
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(exit), 0);
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(mmap), 0);
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(mq_send), 0);
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(mq_receive), 0);

    // Load the seccomp fliter
    seccomp_load(ctx);
}

void start_app(char* filepath) {
    int pid = fork();
    
    // Not a child process
    if(pid != 0){
        return;
    }
    
    mqd_t messageq = mq_open(filepath, O_RDWR | O_CREAT)
}

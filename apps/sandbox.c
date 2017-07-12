#include <seccomp.h> /* libseccomp */

/*
TODO: Enable CGroups for usage limitations and namespaces for isolation
*/

void seccomp_filtered_proc() {
    // ensure none of our children will ever be granted more priv (via setuid, capabilities, ...)
    prctl(PR_SET_NO_NEW_PRIVS, 1);
    // ensure no escape is possible via ptrace
    prctl(PR_SET_DUMPABLE, 0);

    // Init the seccomp filter
    scmp_filter_ctx ctx;
    ctx = seccomp_init(SCMP_ACT_KILL);

    // Allow read, write(both limited by namespaces), exit, sigreturn
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(rt_sigreturn), 0);
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(exit), 0);
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(read), 0);
    seccomp_rule_add(ctx, SCMP_ACT_ALLOW, SCMP_SYS(write), 0);

    // Load the seccomp fliter
    seccomp_load(ctx);
}
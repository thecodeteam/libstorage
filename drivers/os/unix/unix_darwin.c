// +build darwin

#include "unix_darwin.h"
#include <errno.h>

statfs_result _statfs(char* path) {
    statfs_result r;
    r.val = (struct statfs*)malloc(sizeof(struct statfs));
    r.err = statfs((const char*) path, r.val) == 0 ? 0 : errno;
    return r;
}

getmntinfo_result _getmntinfo(int flags) {
    getmntinfo_result r;
    r.err = 0;
    r.len = getmntinfo(&r.val, flags);
    if (r.len < 1) {
        r.err = errno;
        r.len = 0;
    }
    return r;
}

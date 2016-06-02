// +build darwin

#include <stdlib.h>
#include <sys/param.h>
#include <sys/mount.h>

typedef struct {
    struct statfs*      val;
    int                 err;
} statfs_result;

statfs_result _statfs(char* path);

typedef struct {
    int                 len;
    struct statfs*      val;
    int                 err;
} getmntinfo_result;

getmntinfo_result _getmntinfo(int flags);

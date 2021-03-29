#include <stdio.h>
#include <utmp.h>
#include <errno.h>
#include <string.h>
#include "impl.h"

unsigned int getlast_start(const char *path) {
    if(path == 0 || path[0] == '\0') {
        path = "/var/log/wtmp";
    }
    if (utmpname(path) == -1) {
        return 0;
    }
    setutent();
    return 1;
}

struct LastItem* getlast() {
    struct utmp *ut = 0;
    static struct LastItem item;
    ut = getutent();
    if (ut == 0) {
        return 0;
    }
    memset(&item, 0, sizeof(item));
    item.ut_type = ut->ut_type;
    item.ut_pid = ut->ut_pid;
    item.ut_time = (int)((ut->ut_tv).tv_sec);
    memcpy(item.ut_line, ut->ut_line, sizeof(item.ut_line));
    memcpy(item.ut_user, ut->ut_user, sizeof(item.ut_user));
    memcpy(item.ut_host, ut->ut_host, sizeof(item.ut_host));
    return &item;
}

void getlast_end() {
    endutent();
}
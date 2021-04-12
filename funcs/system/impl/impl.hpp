#ifndef _impl_header_h
#define _impl_header_h

#include <stdint.h>
#include <unistd.h>

struct LastItem {
    int ut_type;
    int ut_pid;
    char ut_line[32];
    char ut_user[32];
    char ut_host[256];
    int ut_time;
};

unsigned int getlast_start(const char *path);
struct LastItem* getlast();
void getlast_end();
#endif
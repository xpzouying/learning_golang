#include "demo.h"

CFriendList NewCFriendList(int cnt) {
    CFriend* friends = new CFriend[cnt];
    for (int i = 0; i < cnt; ++i) {
        friends[i].id = i;
        friends[i].age = 20+i;
    }

    CFriendList list;
    list.friends = friends;
    list.length = cnt;

    return list;
}


void DeleteCFriendList(struct CFriendList list) {
    if (list.friends == 0) {
        return;
    }

    delete[] list.friends;
    list.friends = 0;
    return;
}

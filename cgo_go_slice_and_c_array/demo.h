#ifndef _ZY_DEMO_H_
#define _ZY_DEMO_H_

#ifdef __cplusplus
extern "C" {
#endif

typedef struct CFriend {
    int id;
    int age;
} CFriend;

typedef struct CFriendList {
    CFriend *friends;
    int length;
} CFriendList;


CFriendList NewCFriendList(int cnt);
void DeleteCFriendList(struct CFriendList list);


#ifdef __cplusplus
}
#endif

#endif
package main

// #include "demo.h"
import "C"
import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"unsafe"
)

type Friend struct {
	ID  int
	Age int
}

// ----- Go Object to C Object -----
func toCFriend(f Friend) C.CFriend {
	return C.CFriend{
		id:  C.int(f.ID),
		age: C.int(f.Age),
	}
}

// ----- C Object to Go Object -----
func toGoFriend(cf C.CFriend) Friend {
	return Friend{
		ID:  int(cf.id),
		Age: int(cf.age),
	}
}

// --- Go Slice to C array ---
func toCFriends(friends []Friend) (*C.CFriendList, error) {
	l := len(friends)

	if l == 0 {
		return nil, errors.New("empty friend list")
	}

	cFriends := make([]C.CFriend, l)
	for i, f := range friends {
		cFriends[i] = C.CFriend{
			id:  C.int(f.ID),
			age: C.int(f.Age),
		}
	}

	return &C.CFriendList{
		friends: (*C.CFriend)(&cFriends[0]),
		length:  C.int(l),
	}, nil
}

// --- C array to Go Slice: 1 ---
func toGoFriends(cFriendList C.CFriendList) []Friend {
	cFriends := cFriendList.friends
	length := int(cFriendList.length)

	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cFriends)),
		Len:  length,
		Cap:  length,
	}
	slice := *((*[]C.CFriend)(unsafe.Pointer(&hdr)))

	goFriends := make([]Friend, length)
	for i, cf := range slice {
		goFriends[i] = Friend{ID: int(cf.id), Age: int(cf.age)}
	}

	return goFriends
}

// --- C array to Go Slice: 2 ---
func toGoFriends2(cFriendList C.CFriendList) []Friend {
	cFriends := (*[1 << 30]C.CFriend)(unsafe.Pointer(cFriendList.friends))
	length := int(cFriendList.length)

	goFriends := make([]Friend, length)
	for i := 0; i < length; i++ {
		cf := cFriends[i]

		goFriends[i] = Friend{ID: int(cf.id), Age: int(cf.age)}
	}

	return goFriends
}

func main() {
	{ // --- go friend && c friend ---
		f1 := Friend{ID: 1, Age: 20}
		cf1 := toCFriend(f1)
		if err := equalFriend(f1, cf1); err != nil {
			log.Printf("go friend != c friend: %v", err)
		} else {
			log.Printf("toCFriend() succ")
		}
	}

	{ // --- c friend && go friend ---
		cf := C.CFriend{id: 1, age: 20}

		f := toGoFriend(cf)
		if err := equalFriend(f, cf); err != nil {
			log.Printf("go friend != c friend: %v", err)
		} else {
			log.Printf("toGoFriend() succ")
		}
	}

	{ // --- go friend slice to c array ---
		cnt := 10
		friends := newFriends(cnt)
		cFriendList, err := toCFriends(friends)
		if err != nil {
			log.Printf("toCFriends() error: %v", err)
			return
		}

		if err := equalFriendList(friends, *cFriendList); err != nil {
			log.Printf("toCFriends() error: %v", err)
		} else {
			log.Printf("toCFriends() succ")
		}
	}

	{ // --- c array to go slice ---
		cnt := 10
		cFriendList := C.NewCFriendList(C.int(cnt))
		defer C.DeleteCFriendList(cFriendList)

		goFriends := toGoFriends(cFriendList)

		if err := equalFriendList(goFriends, cFriendList); err != nil {
			log.Printf("toGoFriends() error: %v", err)
		} else {
			log.Printf("toGoFriends() succ")
		}
	}

	{ // --- c array to go slice ---
		cnt := 10
		cFriendList := C.NewCFriendList(C.int(cnt))
		defer C.DeleteCFriendList(cFriendList)

		goFriends := toGoFriends2(cFriendList)

		if err := equalFriendList(goFriends, cFriendList); err != nil {
			log.Printf("toGoFriends2() error: %v", err)
		} else {
			log.Printf("toGoFriends2() succ")
		}
	}
}

// new friends, index begin with 0
func newFriends(cnt int) []Friend {
	friends := make([]Friend, cnt)

	for i := 0; i < cnt; i++ {
		friends[i] = Friend{ID: i, Age: 20 + i}
	}

	return friends
}

func equalFriend(f Friend, cf C.CFriend) error {
	goID, cID := f.ID, cf.id
	if goID != int(cID) {
		fmt.Errorf("go id=%d c id=%d", goID, cID)
	}

	goAge, cAge := f.Age, cf.age
	if goAge != int(cAge) {
		fmt.Errorf("go age=%d c age=%d", goAge, cAge)
	}

	return nil
}

func equalFriendList(goFriends []Friend, cFriendList C.CFriendList) error {
	goLen, cLen := len(goFriends), cFriendList.length
	if goLen != int(cLen) {
		return errors.New("list length not equal")
	}

	if goLen == 0 {
		return errors.New("list is empty")
	}

	cFriends := (*[1 << 30]C.CFriend)(unsafe.Pointer(cFriendList.friends))

	for i := 0; i < goLen; i++ {
		gof, cf := goFriends[i], cFriends[i]

		if err := equalFriend(gof, cf); err != nil {
			log.Printf("equalFriendList() not equal: index=%d, %v", i, err)
			return errors.New("list not equal")
		}
	}

	return nil
}

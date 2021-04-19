package main

import "log"

type testStruct struct{}

func NilOrNot(v interface{}) bool {
	return v == nil
}

func main() {
	var s *testStruct

	log.Println("s==nil: ", s == nil)
	log.Println("NilOrNot(interface{}): ", NilOrNot((s)))
}

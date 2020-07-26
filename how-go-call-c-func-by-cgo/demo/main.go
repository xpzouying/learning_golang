package main

/*
typedef struct person {
	char* name;
	int score1;
	int score2;
} person;

person get_person() {
	person zy;
	zy.name = "zouying";
	zy.score1 = 100;
	zy.score2 = 100;

	return zy;
}

int sum(int a, int b) { return a+b; }
*/
import "C"

import "log"

func SayHello() { println("hello ZOUYING") }

func main() {

	SayHello()

	p, err := C.get_person()
	log.Printf("%#v, size of person: %d, err=%v", p, C.sizeof_struct_person, err)

	value := C.sum(p.score1, p.score2)
	println("score=", value)
}

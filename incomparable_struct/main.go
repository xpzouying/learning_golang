package main

import (
	"log"
	"unsafe"
)

type Incomparable [0]func()

type Person struct {
	_    Incomparable
	Name string
}

func main() {
	p1, p2 := Person{Name: "z1"}, Person{Name: "z2"}

	log.Println(p1 == p2)
	log.Println("sizeof(p1)=", unsafe.Sizeof(p1))
	log.Println("sizeof(p2)=", unsafe.Sizeof(p2))
}

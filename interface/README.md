# `interface`

## 思考题

```go
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
```

**问题：** 答案是什么？


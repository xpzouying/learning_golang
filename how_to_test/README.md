# HOW TO TESTING



## 原始代码

代码功能：访客记次数。

```go
package main

import (
	"fmt"
	"log"
	"net/http"
)

var counter = map[string]int{}

func handleHello(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	counter[name]++

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte("<h1 style='color: " + r.FormValue("color") +
		"'>Welcome!</h1> <p>Name: " + name + "</p> <p>Count: " + fmt.Sprint(counter[name]) + "</p>"))
}

func main() {
	http.HandleFunc("/hello", handleHello)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```



运行：

```bash
go run main.go
```



浏览器访问：

![image-20181222170909680](./assets/image-20181222170909680-5469749.png)

本地日志记录：

![image-20181222171507017](./assets/image-20181222171507017.jpg)


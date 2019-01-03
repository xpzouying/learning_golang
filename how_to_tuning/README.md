# HOW TO TUNING

介绍使用Benchmark进行调优。

- 如何写压测示例；
- 如何使用压测程序进行调优；



## 原始代码

代码功能：访客记次数。

```go
package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/sirupsen/logrus"
)

var counter = map[string]int{}
var mu sync.Mutex // mutex for counter

func handleHello(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	mu.Lock()
	counter[name]++
	cnt := counter[name]
	mu.Unlock()

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte("<h1 style='color: " + r.FormValue("color") +
		"'>Welcome!</h1> <p>Name: " + name + "</p> <p>Count: " + fmt.Sprint(cnt) + "</p>"))

	logrus.WithFields(logrus.Fields{
		"module": "main",
		"name":   name,
		"count":  cnt,
	}).Infof("visited")
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	http.HandleFunc("/hello", handleHello)
	logrus.Fatal(http.ListenAndServe(":8080", nil))
}
```



上一次讲了如何进行普通的测试，这次对`handleHello`处理函数编写压力测试示例。

```go
func BenchmarkHandleFunc(b *testing.B) {
	logrus.SetOutput(ioutil.Discard)  // 抛弃日志

	rw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/hello?name=zouying", nil)

	for i := 0; i < b.N; i++ {
		handleHello(rw, req)
	}
}
```



运行压测用例：

```bash
➜  how_to_tuning git:(master) ✗ go test -bench .
goos: darwin
goarch: amd64
BenchmarkHandleFunc-8             300000              4116 ns/op
PASS
ok      _/Users/zouying/src/Github.com/ZOUYING/learning_golang/how_to_tuning    1.319s
```



或者增加`-benchmem`选项，显示内存信息，

```bash
➜  how_to_tuning git:(master) ✗ go test -bench . -benchmem
goos: darwin
goarch: amd64
BenchmarkHandleFunc-8             300000              4297 ns/op            1411 B/op         25 allocs/op
PASS
ok      _/Users/zouying/src/Github.com/ZOUYING/learning_golang/how_to_tuning    1.368s
```


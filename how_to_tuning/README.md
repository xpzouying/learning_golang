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

或者在测试用例中的最开始，增加下列代码，也可以显示内存信息。

```go
b.ReportAllocs()
```



每个压测用例默认的压测时常大概在1秒钟，如果我们需要压测的时间长一些的话，那么可以在运行的时候，加上`-benchtime=5s`的参数，5s表示5秒。



## Golang调优

### 背景

- Robert Hundt在2011年Scala Day发表了一篇论文，论文叫：[Loop Recognition in C++/Java/Go/Scala](https://ai.google/research/pubs/pub37122)，大概讲的就是论文中的Go程序运行的非常慢。

- Go团队就使用`go tool pprof`进行了优化，具体参见：[profiling-go-programs](https://blog.golang.org/profiling-go-programs)。结果为：

  - 速度巨大提升（*magnitude faster*）
  - 6倍的内存降低（*use 6x less memory*）

- 重现论文程序。在论文中虽然比对了四种语言，但由于go团队的人没有足够的Java和Scala的优化能力，所以只进行了Go和C++具体比对。

  >  具体的软件、硬件如下：
  >
  > ```bash
  > $ go version
  > go version devel +08d20469cc20 Tue Mar 26 08:27:18 2013 +0100 linux/amd64
  > $ g++ --version
  > g++ (GCC) 4.8.0
  > Copyright (C) 2013 Free Software Foundation, Inc.
  > ...
  > $
  > ```
  >
  > ```
  > 硬件：
  > 
  > 3.4GHz Core i7-2600 CPU and 16 GB of RAM running Gentoo Linux's 3.8.4-gentoo kernel
  > ```
  >
  > **调优前，重现论文程序的结果。**
  >
  > 具体结果：

  > ```bash
  > $ cat xtime
  > #!/bin/sh
  > /usr/bin/time -f '%Uu %Ss %er %MkB %C' "$@"
  > $
  > 
  > $ make havlak1cc
  > g++ -O3 -o havlak1cc havlak1.cc
  > $ ./xtime ./havlak1cc
  > # of loops: 76002 (total 3800100)
  > loop-0, nest: 0, depth: 0
  > 17.70u 0.05s 17.80r 715472kB ./havlak1cc
  > $
  > 
  > $ make havlak1
  > go build havlak1.go
  > $ ./xtime ./havlak1
  > # of loops: 76000 (including 1 artificial root node)
  > 25.05u 0.11s 25.20r 1334032kB ./havlak1
  > $
  > ```
  >
  > 输出参数为：`u: user time`, `s: system time`,`r: real time`
  >
  > C++程序运行了17.80s，使用内存700MB；
  >
  > Go程序是运行了25.20s，使用内存1302MB；

  - 



### Golang自带的调优库: pprof

- runtime/pprof：输出runtime的profiling数据，写到指定文件中，而该文件可以被一些pprof tool打开。

  - 使用benchmark

    > ```bash
    > # 输出cpu profile和memory profile
    > go test -cpuprofile cpu.prof -memprofile mem.prof -bench .
    > ```

  - 标准的程序

    > ```go
    > var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to `file`")
    > var memprofile = flag.String("memprofile", "", "write memory profile to `file`")
    > 
    > func main() {
    >     flag.Parse()
    >     if *cpuprofile != "" {
    >         f, err := os.Create(*cpuprofile)
    >         if err != nil {
    >             log.Fatal("could not create CPU profile: ", err)
    >         }
    >         if err := pprof.StartCPUProfile(f); err != nil {
    >             log.Fatal("could not start CPU profile: ", err)
    >         }
    >         defer pprof.StopCPUProfile()
    >     }
    > 
    >     // ... rest of the program ...
    > 
    >     if *memprofile != "" {
    >         f, err := os.Create(*memprofile)
    >         if err != nil {
    >             log.Fatal("could not create memory profile: ", err)
    >         }
    >         runtime.GC() // get up-to-date statistics
    >         if err := pprof.WriteHeapProfile(f); err != nil {
    >             log.Fatal("could not write memory profile: ", err)
    >         }
    >         f.Close()
    >     }
    > }
    > ```

- net/http/pprof：也可以通过HTTP server获取到runtime profiling data，一般如果是http服务的话，可以直接挂在到对应的http handler上，然后通过访问`/debug/pprof/`开头的路径，进行进行相应数据的访问。

  > ```go
  > import _ "net/http/pprof"
  > 
  > go func() {
  > 	log.Println(http.ListenAndServe("localhost:6060", nil))
  > }()
  > ```
  >
  > ```bash
  > # 查看heap profile
  > go tool pprof http://localhost:6060/debug/pprof/heap
  > 
  > # 查看30s CPU profile
  > go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
  > 
  > # 检查blocking profile，需要设置首先调用runtime.SetBlockProfileRate
  > go tool pprof http://localhost:6060/debug/pprof/block
  > 
  > # 收集5s的trace数据
  > wget http://localhost:6060/debug/pprof/trace?seconds=5
  > ```

* 打开收集到的profile

  * golang pprof tool

    ```bash
    go tool pprof cpu.prof
    ```

  * 第三方的pprof tool

    * uber火焰图



如何使用工具：

- topN



## CPU调优

原理：

每秒钟100次的数据状态采样，





启动压测时，我们加入`-cpuprofile`参数选项，可以生成

```bash
➜  how_to_tuning git:(master) ✗ go test -bench . -cpuprofile=/tmp/cpu.prof
goos: darwin
goarch: amd64
BenchmarkHandleFunc-8             300000              5105 ns/op            1414 B/op    25 allocs/op
PASS
ok      _/Users/zouying/src/Github.com/ZOUYING/learning_golang/how_to_tuning    2.772s
```



打开生成的cpu profile文件，使用`topN`，或者使用`topN -cum`：

```bash
➜  how_to_tuning git:(master) ✗ go tool pprof /tmp/cpu.prof
Type: cpu
Time: Jan 12, 2019 at 11:12pm (CST)
Duration: 2.74s, Total samples = 2.29s (83.46%)
Entering interactive mode (type "help" for commands, "o" for options)
(pprof) top10
Showing nodes accounting for 890ms, 38.86% of 2290ms total
Dropped 54 nodes (cum <= 11.45ms)
Showing top 10 nodes out of 124
      flat  flat%   sum%        cum   cum%
     140ms  6.11%  6.11%      140ms  6.11%  runtime.memclrNoHeapPointers
     130ms  5.68% 11.79%      670ms 29.26%  runtime.mallocgc
     120ms  5.24% 17.03%      290ms 12.66%  runtime.mapassign_faststr
     110ms  4.80% 21.83%      180ms  7.86%  time.Time.AppendFormat
      80ms  3.49% 25.33%       80ms  3.49%  runtime.kevent
      80ms  3.49% 28.82%      100ms  4.37%  runtime.mapiternext
      70ms  3.06% 31.88%       70ms  3.06%  runtime.stkbucket
      60ms  2.62% 34.50%     1070ms 46.72%  github.com/sirupsen/logrus.(*TextFormatter).Format
      50ms  2.18% 36.68%       50ms  2.18%  cmpbody
      50ms  2.18% 38.86%       80ms  3.49%  runtime.heapBitsSetType
(pprof)


(pprof) top 10 -cum
Showing nodes accounting for 0.26s, 11.35% of 2.29s total
Dropped 54 nodes (cum <= 0.01s)
Showing top 10 nodes out of 124
      flat  flat%   sum%        cum   cum%
         0     0%     0%      2.12s 92.58%  _/Users/zouying/src/Github.com/ZOUYING/learning_golang/how_to_tuning.BenchmarkHandleFunc
     0.02s  0.87%  0.87%      2.12s 92.58%  _/Users/zouying/src/Github.com/ZOUYING/learning_golang/how_to_tuning.handleHello
         0     0%  0.87%      2.12s 92.58%  testing.(*B).launch
         0     0%  0.87%      2.12s 92.58%  testing.(*B).runN
     0.01s  0.44%  1.31%      1.39s 60.70%  github.com/sirupsen/logrus.(*Entry).Infof
     0.02s  0.87%  2.18%      1.28s 55.90%  github.com/sirupsen/logrus.(*Entry).Info
     0.02s  0.87%  3.06%      1.23s 53.71%  github.com/sirupsen/logrus.Entry.log
         0     0%  3.06%      1.10s 48.03%  github.com/sirupsen/logrus.(*Entry).write
     0.06s  2.62%  5.68%      1.07s 46.72%  github.com/sirupsen/logrus.(*TextFormatter).Format
     0.13s  5.68% 11.35%      0.67s 29.26%  runtime.mallocgc
```



运行`list handleHello`查看handleHello函数的状态：

```bash
(pprof) list handleHello
Total: 2.29s
ROUTINE ======================== _/Users/zouying/src/Github.com/ZOUYING/learning_golang/how_to_tuning.handleHello in /Users/zouying/src/Github.com/ZOUYING/learning_golang/how_to_tuning/main.go
      20ms      2.12s (flat, cum) 92.58% of Total
         .          .     12:var mu sync.Mutex // mutex for counter
         .          .     13:
         .          .     14:func handleHello(w http.ResponseWriter, r *http.Request) {
         .          .     15:   name := r.FormValue("name")
         .          .     16:   mu.Lock()
         .       20ms     17:   counter[name]++
         .          .     18:   cnt := counter[name]
         .       10ms     19:   mu.Unlock()
         .          .     20:
         .       70ms     21:   w.Header().Set("Content-Type", "text/html; charset=utf-8")
         .      120ms     22:   w.Write([]byte("<h1 style='color: " + r.FormValue("color") +
      10ms       70ms     23:           "'>Welcome!</h1> <p>Name: " + name + "</p> <p>Count: " + fmt.Sprint(cnt) + "</p>"))
         .          .     24:
         .      390ms     25:   logrus.WithFields(logrus.Fields{
      10ms       10ms     26:           "module": "main",
         .       30ms     27:           "name":   name,
         .       10ms     28:           "count":  cnt,
         .      1.39s     29:   }).Infof("visited")
         .          .     30:}
         .          .     31:
         .          .     32:func main() {
         .          .     33:   logrus.SetFormatter(&logrus.JSONFormatter{})
         .          .     34:
```





## 参考

- [golang/pprof](https://golang.org/pkg/runtime/pprof/)
- [golang/profiling-go-programs](https://blog.golang.org/profiling-go-programs)
- [Google 推出 C++ Go Java Scala的基准性能测试](https://www.cnbeta.com/articles/soft/145252.htm)
- 
# single flight

当相同请求被请求多次时，只有一次请求会在发送的过程中，此时若有新的请求到达时，等待已经发送的请求结果。

**举例说明：**

我们发出5个并发，同时请求http://163.com。

若使用singleflight，只有第1次会真正的发送给http://163.com，另外4个请求阻塞等待第1次请求的结果。


### 运行

```bash
go run main.go
```


**结果：**

`do http request`为真正发出的http请求，

由此看来，真实http request只发生过一次。

```bash
2020/02/21 16:05:00 do http request
2020/02/21 16:05:00 goroutine-4, result: finish
2020/02/21 16:05:00 goroutine-3, result: finish
2020/02/21 16:05:00 goroutine-0, result: finish
2020/02/21 16:05:00 goroutine-2, result: finish
2020/02/21 16:05:00 goroutine-1, result: finish
```
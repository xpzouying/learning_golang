# json benchmark

使用图片Base64，约160KB，进行测试。

### 对比Umarshal

对比库：

- fastjson
- stdjson

速度对比：

**fastjson**

运行：

```bash
go test -benchmem -bench=BenchmarkImageBase64_fastjson$
```

结果：

```
goos: darwin
goarch: amd64
pkg: github.com/xpzouying/learning_golang/json_marshal_benchmark
BenchmarkImageBase64_fastjson-8            20000             65925 ns/op          294927 B/op          1 allocs/op
PASS
ok      github.com/xpzouying/learning_golang/json_marshal_benchmark     2.077s
```


**stdjson**

运行：

```bash
go test -benchmem -bench=BenchmarkImageBase64_stdjson$
```

结果：

```
goos: darwin
goarch: amd64
pkg: github.com/xpzouying/learning_golang/json_marshal_benchmark
BenchmarkImageBase64_stdjson-8               500           3267756 ns/op          221387 B/op          4 allocs/op
PASS
ok      github.com/xpzouying/learning_golang/json_marshal_benchmark     2.515s
```


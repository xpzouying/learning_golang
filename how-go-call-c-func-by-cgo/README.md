# [UnderHood] Go如何通过cgo调用c函数（1）

## 前言

最近在工作中经常会用到cgo，在此之间也遇到了不少的问题，也有不少人问我cgo底层是怎么工作的，所以在此把我自己的疑问和别人问我的问题汇总一下。

准备把cgo这部分文件分为多篇完成，由浅至深。

### 常见问题

- [ ] cgo是如何工作的？Go是如何调用起来C的代码呢？
- [ ] cgo的函数调用跟普通的函数调用有什么区别？
- [ ] go stack和c stack会不会出现问题？
- [ ] cgo是不是慢？为什么？



## CGO：C和Go是如何一起工作的？

根据上面的问题，慢慢剖析cgo吧。以下只是个人理解，可能有误。

### 环境介绍

**编译环境：**

```bash
➜  learning_golang ✗ go version       
go version go1.14.3 darwin/amd64
```

### 基本概念

对于操作系统来说，无论是C、Go编写的函数，无非是一堆指令，最终对于操作系统来说都是一堆指令。CPU根据指令顺序执行。对于函数的调用，简单的理解为跳到某个指令段，然后开始顺序执行对应的指令。

<img src="image-20200726162130770.png" alt="image-20200726162130770" style="zoom:50%;" />

[图片参考：How CPU WORKS](https://www.youtube.com/watch?v=cNN_tTXABUA)

在Go中，调用函数就只需要告诉Go：

2. 哪一个函数：函数的地址
2. 参数的地址：由于多个参数也是连续存放的，所以取参数的首地址。

Go通过cgo也是跳不出这套规则。那么，cgo对C的函数是如何处理的，可以让Go调用起来C函数呢？

### 代码分析

#### CGO DEMO

从cgo demo开始，具体代码如下：

```go
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

	p := C.get_person()
	log.Printf("%#v, size of person: %d", p, C.sizeof_struct_person)

	value := C.sum(p.score1, p.score2)
	println("score=", value)
}
```

在代码中，

- C语言部分：
  - 定义了1个结构体person
  - 定义了2个函数：其中一个有输入参数，一个没有输入参数
  - C相关的代码必须定义在注释中，可以包含：函数、变量的声明和定义。
  - 函数、变量名可以想象成被定义在一个叫`"C"`的package中。

- Go语言部分：
  - `import "C"`跟上面C的代码不能有空行。否则会报错：`./main.go:18:2: could not determine kind of name for C.get_person`。
  - 使用`import "C"`。"C"并不是Go真实的一个package，是让Go能调用C的符号，比如`C.int`、`C.sum()`等。
  - 定义了1个Go函数。

更多的cgo规范细节可以参考官方文档：

- [cmd/cgo/doc.go](https://golang.org/src/cmd/cgo/doc.go)

#### 运行

```bash
hello ZOUYING
2020/07/26 19:05:06 main._Ctype_struct_person{name:(*main._Ctype_char)(0x416cebc), score1:100, score2:100}, size of person: 16
score= 200                                                                                                                    
```

从结果中，通过对`C.get_person()`返回值打印的日志，可以看到`C.person`类型变成了`main._Ctype_struct_person`，那么在此过程中Go编译器又做了哪些工作呢？

#### 生成中间文件

运行命令`go tool cgo main.go`生成cgo中间文件。在当前文件夹中会生成`_obj`的文件夹，保存cgo的中间文件。具体的生成过程可以参考`$GOROOT/src/cmd/cgo/doc.go`，在此先不赘述。

```bash
(base) ➜  demo (how-go-call-c-func-by-cgo) ✗ ls -lh _obj 
total 64
-rw-r--r--  1 zouying  staff   3.3K Jul 26 17:44 _cgo_.o
-rw-r--r--  1 zouying  staff   605B Jul 26 17:44 _cgo_export.c
-rw-r--r--  1 zouying  staff   1.5K Jul 26 17:44 _cgo_export.h
-rw-r--r--  1 zouying  staff    13B Jul 26 17:44 _cgo_flags
-rw-r--r--  1 zouying  staff   1.8K Jul 26 17:44 _cgo_gotypes.go
-rw-r--r--  1 zouying  staff   416B Jul 26 17:44 _cgo_main.c
-rw-r--r--  1 zouying  staff   421B Jul 26 17:44 main.cgo1.go
-rw-r--r--  1 zouying  staff   2.7K Jul 26 17:44 main.cgo2.c
```

#### 生成中间文件的流程

Go在产生这些中间文件时，会经过一系列工作。主要流程包括，

1. 分析C的代码：借助gcc分析当前cgo的所有标识符、判断是否存在错误等等。
2. 将C翻译成Go的代码：根据main.go的内容转换成中间的go文件，即产生了上面列表中的.go、.c、.h文件。
3. 其他：在此过程中还会生成一些链接库，在此就先跳过，详细的细节可以参考`cgo/doc.go`。

#### 分析中间文件

分析一下_obj这个临时文件夹中的文件。

首先，main.go会被拷贝成`main.cgo1.go`文件，并且在此过程中会把对应的C函数进行翻译。

该文件删除大部分注释后，源码如下，

```go
package main

import _ "unsafe"

import "log"

func SayHello() { println("hello ZOUYING") }

func main() {

	SayHello()

	p, err := ( /*line :31:12*/_C2func_get_person /*line :31:23*/)()
	log.Printf("%#v, size of person: %d, err=%v", p, ( /*line :32:51*/_Ciconst_sizeof_struct_person /*line :32:72*/), err)

	value := ( /*line :34:11*/_Cfunc_sum /*line :34:15*/)(p.score1, p.score2)
	println("score=", value)
}
```

在`main.cgo1.go`中可以看到部分的翻译迹象，类似于：`C.get_person`变成了`_C2func_get_person`、`C.sum`变成了`_Cfunc_sum`等。这些翻译后的定义在`_cgo_gotypes.go`文件中。



**C.person的定义：**

看`C.get_person()`。该方法没有输入，只有返回值。该函数有2个返回值，一个是C.person，一个是error，其中第二个返回值是可选的，编译器会根据是否需要第二个返回值然后定义翻译后方法的返回值个数，可以对比`C.get_person`和`C.sum`翻译后的区别。

```go
//go:cgo_unsafe_args
func _C2func_get_person() (r1 _Ctype_struct_person, r2 error) {
	errno := _cgo_runtime_cgocall(_cgo_299c25848d85_C2func_get_person, uintptr(unsafe.Pointer(&r1)))
	if errno != 0 { r2 = syscall.Errno(errno) }
	if _Cgo_always_false {
	}
	return
}
```

该函数其实就是通过`_cgo_runtime_cgocall`对函数进行调用。该函数的定义为：

```go
//go:linkname _cgo_runtime_cgocall runtime.cgocall
func _cgo_runtime_cgocall(unsafe.Pointer, uintptr) int32
```

`//go:xxx`表示编译器指令，编译器接收注释形式的指令。注意，为了区分与正常注释，`//`和`go`之间不能有空格。

`go:linkname`编译标志会将`_cgo_runtime_cgocall`链接到`runtime.cgocall`，输入为1、函数地址，2、参数首地址，返回值为调用编号，返回0表示调用成功。`runtime.cgocall`后续文章再详细分析，再此先不展开。

传入参数包括：函数名`_cgo_299c25848d85_C2func_get_person`和返回值的地址。对于通用的函数调用时，输入应该是函数和args列表的首地址，在此直接取返回值的地址是由于没有输入参数。

其中函数的定义为：

```go
//go:cgo_import_static _cgo_299c25848d85_C2func_get_person
//go:linkname __cgofn__cgo_299c25848d85_C2func_get_person _cgo_299c25848d85_C2func_get_person
var __cgofn__cgo_299c25848d85_C2func_get_person byte
var _cgo_299c25848d85_C2func_get_person = unsafe.Pointer(&__cgofn__cgo_299c25848d85_C2func_get_person)
```

`_cgo_299c25848d85_C2func_get_person`被定义为取`__cgofn__cgo_299c25848d85_C2func_get_person`的地址，而该方法又被编译标志链接到`_cgo_299c25848d85_C2func_get_person`方法，该方法的定义在`main.cgo2.c`中，定义如下，

```c
CGO_NO_SANITIZE_THREAD
int
_cgo_299c25848d85_C2func_get_person(void *v)
{
	int _cgo_errno;
	struct {
		person r;
	} __attribute__((__packed__)) *_cgo_a = v;
	char *_cgo_stktop = _cgo_topofstack();
	__typeof__(_cgo_a->r) _cgo_r;
	_cgo_tsan_acquire();
	errno = 0;
	_cgo_r = get_person();
	_cgo_errno = errno;
	_cgo_tsan_release();
	_cgo_a = (void*)((char*)_cgo_a + (_cgo_topofstack() - _cgo_stktop));
	_cgo_a->r = _cgo_r;
	_cgo_msan_write(&_cgo_a->r, sizeof(_cgo_a->r));
	return _cgo_errno;
}
```

该方法中会最终实际调用到我们定义的C方法：`get_person()`，而`get_person`函数的定义同样被复制在`main.cgo2.c`中。



**C.sum的定义：**

```go
//go:cgo_import_static _cgo_299c25848d85_Cfunc_sum
//go:linkname __cgofn__cgo_299c25848d85_Cfunc_sum _cgo_299c25848d85_Cfunc_sum
var __cgofn__cgo_299c25848d85_Cfunc_sum byte
var _cgo_299c25848d85_Cfunc_sum = unsafe.Pointer(&__cgofn__cgo_299c25848d85_Cfunc_sum)

//go:cgo_unsafe_args
func _Cfunc_sum(p0 _Ctype_int, p1 _Ctype_int) (r1 _Ctype_int) {
	_cgo_runtime_cgocall(_cgo_299c25848d85_Cfunc_sum, uintptr(unsafe.Pointer(&p0)))
	if _Cgo_always_false {
		_Cgo_use(p0)
		_Cgo_use(p1)
	}
	return
}
```

该函数大致流程都与`C.person`方法一致，不同的在于有对应的输入、输出参数，所以在对其调用的时候，是把第一个入参的地址作为`_cgo_runtime_cgocall`的第2输入参数传入。



**总结**

总结一下cgo的调用关系：

<img src="image-20200728151755957.png" alt="image-20200728151755957" style="zoom:50%;" />


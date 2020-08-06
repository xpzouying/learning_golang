# [CGO系列]GO与CGO对象的最佳实践

## 前言

在开发过程中，常常遇到Go对象与cgo对象相互转换的情况，在此记录使用过程中的常遇到的问题。

Go和C分别有一个结构体类型，演示Go和C对应的数据结构之间的转换。

Go结构体：

```go
type Friend struct {
	ID  int
	Age int
}
```

C结构体：

```c
typedef struct CFriend {
    int id;
    int age;
} CFriend;

typedef struct CFriendList {
    CFriend *friends;
    int length;
} CFriendList;
```

CFriendList为CFriend数组的封装。

## 一、结构体之间的转换

**1、Go的结构体`Friend`转换为C的结构体数据`CFriend`：**

```go
// ----- Go Object to C Object -----
func toCFriend(f Friend) C.CFriend {
	return C.CFriend{
		id:  C.int(f.ID),
		age: C.int(f.Age),
	}
}
```

测试结果：

```go
f1 := Friend{ID: 1, Age: 20}
cf1 := toCFriend(f1)
```



**2、C结构体`CFriend`转换为Go结构体`Friend`：**

```go
// ----- C Object to Go Object -----
func toGoFriend(cf C.CFriend) Friend {
	return Friend{
		ID:  int(cf.id),
		Age: int(cf.age),
	}
}
```

测试结果：

```go
cf := C.CFriend{id: 1, age: 20}
f := toGoFriend(cf)
```



## 二、数组之间的转换

**1、Go的数组转换为C的数组：**

在这里使用`CFriendList`结构体对C数组进行了封装。

```c
typedef struct CFriendList {
    CFriend *friends;
    int length;
} CFriendList;
```

转换函数：

```go
// --- Go Slice to C array ---
func toCFriends(friends []Friend) (*C.CFriendList, error) {
	l := len(friends)

	if l == 0 {
		return nil, errors.New("empty friend list")
	}

	cFriends := make([]C.CFriend, l)
	for i, f := range friends {
		cFriends[i] = C.CFriend{
			id:  C.int(f.ID),
			age: C.int(f.Age),
		}
	}

	return &C.CFriendList{
		friends: (*C.CFriend)(&cFriends[0]),
		length:  C.int(l),
	}, nil
}
```

返回的C数组的数据为`[]C.CFriend`类型，为在Go中申请的**连续**的内存块，返回该内存的首地址，以`&cFriends[0]`的形式返回。

**2、C数组转换为Go数组**

**方法1：构建slice header**

```go
// --- C array to Go Slice ---
func toGoFriends(cFriendList C.CFriendList) []Friend {
	cFriends := cFriendList.friends
	length := int(cFriendList.length)

	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(cFriends)),
		Len:  length,
		Cap:  length,
	}
	slice := *((*[]C.CFriend)(unsafe.Pointer(&hdr)))

	goFriends := make([]Friend, length)
	for i, cf := range slice {
		goFriends[i] = Friend{ID: int(cf.id), Age: int(cf.age)}
	}

	return goFriends
}
```

首先，需要访问C数组的所有元素，在这里使用的一种方法是，

1. 先利用`C.CFriendList`已有的内存块构建一个Go认知的slice header
2. 然后遍历该slice访问内存中的元素

其次，通过第一步遍历的数据，构建需要返回的Go slice。



参考Go Blog中[slices intro](https://blog.golang.org/slices-intro)文章，对于slice数据结构，内部实现为3个成员：

- ptr：指向数组底层的数据地址
- len：当前数组长度
- cap：当前内存容量大小

![](https://blog.golang.org/slices-intro/slice-struct.png)

以`make([]byte, 5)`为例：

![](https://blog.golang.org/slices-intro/slice-1.png)



**方法2：以指定类型的数组对C内存进行解析**

```go
func toGoFriends2(cFriendList C.CFriendList) []Friend {
	cFriends := (*[1 << 30]C.CFriend)(unsafe.Pointer(cFriendList.friends))
	length := int(cFriendList.length)

	goFriends := make([]Friend, length)
	for i := 0; i < length; i++ {
		cf := cFriends[i]

		goFriends[i] = Friend{ID: int(cf.id), Age: int(cf.age)}
	}

	return goFriends
}
```

对于`(*[1<<30]C.CFriend)`的解释：

1. *表示为指针
2. [1<<30]表示是一个数组。对于slice表达，需要指定一个边界，参考[go spec: slice expressions](https://golang.org/ref/spec#Slice_expressions)，所以在这里用了个极其夸张的边界值。

所以，`(*[1<<30]C.CFriend)`仅仅表示一种类型，这个类型为一个指针，指向一个大小为1<<30大小的数组，数组中的元素为C.CFriend。


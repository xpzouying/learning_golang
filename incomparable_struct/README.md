# 不可比较的struct

定义个类型，是方法，并且长度为0，不占用任何空间。

将该类型放在struct中，让struct不具备比较能力，即==操作符在编译时都失效。

```go
type Incomparable [0]func()
```

```go
type Person struct {
        _    Incomparable
        Name string
}
```

报错：

```
./main.go:18:17: invalid operation: p1 == p2 (struct containing Incomparable cannot be compared)
```
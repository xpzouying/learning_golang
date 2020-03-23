# make slice背后的内存管理


## 介绍

[make的官方文档说明](https://golang.org/pkg/builtin/#make)



## 流程

```go

make([]byte, 4)

```


`make([]byte, 4)`语句会最终转化为：slice.go中的`makeslice`函数，源码文件为：`runtime/slice.go`。

整体流程是：找到一块空间用于存放对象，将该内存空间的地址返回给调用者。

对于go对象的内存管理来说，将对象按照申请的大小分为3类：

1. tiny object：16B以下，且不包含指针的对象
2. small object：32KB以下，或者包含指针的16B以下的对象
3. large object：32KB以上

所以runtime会根据申请的对象类型不同，进行不同的处理。

1. tiny object：直接使用tiny allocator进行内存申请。会复用现有tiny内存块
2. small object：根据sizeclass找到对应的规则进行内存申请
3. large object：直接从heap上进行申请


## 源码解析

具体源码解析如下。

```go
func makeslice(et *_type, len, cap int) unsafe.Pointer {
	mem, overflow := math.MulUintptr(et.size, uintptr(cap))
	if overflow || mem > maxAlloc || len < 0 || len > cap {
		// NOTE: Produce a 'len out of range' error instead of a
		// 'cap out of range' error when someone does make([]T, bignumber).
		// 'cap out of range' is true too, but since the cap is only being
		// supplied implicitly, saying len is clearer.
		// See golang.org/issue/4085.
		mem, overflow := math.MulUintptr(et.size, uintptr(len))
		if overflow || mem > maxAlloc || len < 0 {
			panicmakeslicelen()
		}
		panicmakeslicecap()
	}

	return mallocgc(mem, et, true)
}
```

1. 申请内存空间时，会判断申请的内存空间是否有乘法溢出或者超出申请的限制。
2. 调用`mallocgc`申请内存。入参为，
    - mem：申请大小，单位bytes，大小为`类型*大小`
    - et：类型
    - true：需要置0


**mallocgc**

源码位于：`runtime/malloc.go`

```go
// 为一个对象（object）申请内存，单位：bytes
// 小对象（small objects）从各自的P cache上的空闲列表（free lists）中申请；
// 大对象（Large objects，> 32KB）直接从堆（heap）上申请；
func mallocgc(size uintptr, typ *_type, needzero bool) unsafe.Pointer {
    // 获取当前的m，并标记状态为正在申请内存（mallocing）
	mp := acquirem()
    // ...
	mp.mallocing = 1

	shouldhelpgc := false
	dataSize := size
	c := gomcache()
	var x unsafe.Pointer
    noscan := typ == nil || typ.ptrdata == 0
    
    // maxSmallSize：32768（32 * 1024）
	if size <= maxSmallSize {
        // maxTinySize ： 16
        // 如果申请的大小小于16字节，并且为不包含指针，则使用tiny allocator申请。
        // Tiny allocator的工作原则：
        // 1. 使用1个内存块存放所有的tiny对象；
        // 2. 不包含指针
        // 3. 对象的释放不是显示。如果需要显示地释放对象，申请对象时，确保申请的大小>=16。
        //
        // tiny allocator的主要作用是为：
        // 1. 小字符串
        // 2. 逃逸变量
		if noscan && size < maxTinySize {
			off := c.tinyoffset
            // 申请的内存空间做一定的对齐
			if size&7 == 0 {
				off = alignUp(off, 8)
			} else if size&3 == 0 {
				off = alignUp(off, 4)
			} else if size&1 == 0 {
				off = alignUp(off, 2)
            }
            
            // 如果申请的对象可以放入当前的tiny内存块，则直接使用该内存块
            // 1. 新的对象存放在：地址的起点：c.tiny+off
            // 2. tiny的偏移位更新
            // 3. mcache的tiny计数器递增
            // 4. 释放m的内存申请的状态位
            // 5. 释放m与g的关系
			if off+size <= maxTinySize && c.tiny != 0 {
				// The object fits into existing tiny block.
				x = unsafe.Pointer(c.tiny + off)
				c.tinyoffset = off + size
				c.local_tinyallocs++
				mp.mallocing = 0
				releasem(mp)
				return x
            }

            // 若上面条件不满足，则申请新的tiny内存块
            //
            // mcache中的内存管理是由：数组（134）+链表
            // 获取一个可用的mspan内存块
			span := c.alloc[tinySpanClass]
			v := nextFreeFast(span)
			if v == 0 {
				// 若找不到，则重新申请一份
				v, _, shouldhelpgc = c.nextFree(tinySpanClass)
            }
            // 清空该地址及后面16B的数据
            // - tiny objects为16B内存块，所以需要把该地址及后续16B的内存空间清0
            // - 清0的方法为：
            // 每个uint64为8B，[2]uint64为16B
            // 将该地址标示为：2个uint64的数组空间，分别置这2个元素为0
			x = unsafe.Pointer(v)
			(*[2]uint64)(x)[0] = 0
			(*[2]uint64)(x)[1] = 0
			// See if we need to replace the existing tiny block with the new one
			// based on amount of remaining free space.
			if size < c.tinyoffset || c.tiny == 0 {
				c.tiny = uintptr(x)
				c.tinyoffset = size
			}
			size = maxTinySize
		} else {
			// 对于small object的申请，先根据sizeclass找到合适的mspan，然后在对应的sizeclass链表上申请
			// 若当前mspan没有，则从mcenter/heap上申请新的
			var sizeclass uint8
			if size <= smallSizeMax-8 {
				sizeclass = size_to_class8[(size+smallSizeDiv-1)/smallSizeDiv]
			} else {
				sizeclass = size_to_class128[(size-smallSizeMax+largeSizeDiv-1)/largeSizeDiv]
			}
			size = uintptr(class_to_size[sizeclass])
			spc := makeSpanClass(sizeclass, noscan)
			span := c.alloc[spc]
			v := nextFreeFast(span)
			if v == 0 {
				v, span, shouldhelpgc = c.nextFree(spc)
			}
			x = unsafe.Pointer(v)
			if needzero && span.needzero != 0 {
				memclrNoHeapPointers(unsafe.Pointer(v), size)
			}
		}
	} else {
		// 如果是large object（32KB），则直接从heap上申请
		var s *mspan
		shouldhelpgc = true
		systemstack(func() {
			s = largeAlloc(size, needzero, noscan)
		})
		s.freeindex = 1
		s.allocCount = 1
		x = unsafe.Pointer(s.base())
		size = s.elemsize
	}

	// ...

	return x
}
```
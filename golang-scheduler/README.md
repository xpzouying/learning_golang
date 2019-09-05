从这篇文章开始，通过几篇文章介绍Golang调度的细节。


来源：[个人博客](http://zouying.is/20190725-scheduler1-os/)



在介绍goroutine之前，先回顾操作系统中的一些概念。这也有利于后面介绍Goroutine相应的内容。

### 程序、进程、线程

**什么是程序（Program）**

程序就是一堆指令集代码，计算机通过执行这堆指令集可以完成一系列任务。

**什么是进程（Process）**

一个程序如果被加载到内存，那么它就变成了一个进程（Process），所以进程就是一个运行状态的程序（Program）。进程 = 程序 + 执行。

内存中的进程会被分为4个区域：

![process_components](https://zyblog-1255449766.cos.ap-beijing.myqcloud.com/process_components.jpg)

- stack：主要包含临时变量，例如函数的参数、返回值、本地临时变量
- heap：程序运行中动态申请的空间
- text：包含当前执行指令和处理器中寄存器的内容
- data：包含全局变量、静态（static）变量


**什么是线程（Thread）**


线程就是一段执行代码（a path of execution within a process）+运行时的数据。


线程一般也被称为“轻量级的进程”。线程拥有自己的变量、寄存器值、程序计数器、栈信息，但是与进程中其他的线程共享代码块、数据块和打开的文件。

![](https://zyblog-1255449766.cos.ap-beijing.myqcloud.com/thread_processes.jpg)


**进程和线程的关系**

- 一个进程可以有多个线程
- 进程是分配资源的基本单位；线程是系统调度的最小单位
- 进程拥有完整的资源；线程本身只拥有少量的资源，与同一个进程中的其他线程共享进程中的一些资源，例如：代码段、数据段、打开文件、I/O设备等

> 早期的Linux版本中，1996年，Linus Torvalds提及过，线程和进程都是执行的上下文内容（“both threads and processes are treated as a ‘context of execution’”），其中包括的资源有：CPU状态、MMU状态、permissions、各种communication states（比如open files、signal handlers等等）


**为什么引入线程**

- 线程是轻量级进程，相对于进程来说，创建和终止更为快速；
- 由于线程拥有的资源比较少，线程在切换时，只需要保存更少的资源，这样也就使得线程在做上下文切换时更为快速


### 调度

**什么是调度？**

调度就是将正在运行的进程从CPU上移除，然后从就绪队列中选择一个进程分配CPU资源运行。


**调度器的作用**

操作系统同时有很多任务要运行，任务只有分配到相应的硬件资源后（比如分配到CPU、内存）才能运行，然而可用的硬件资源（CPU、内存等）只有一份，如何给这些任务合理地分配硬件资源，让每个任务都得到合理地运行，这就需要调度器按照一定的策略对任务进行调度。

**Context Switch / 上下文切换**

当进程运行过程中，由于某些原因（如I/O、CPU时间片用完等等），进程从CPU被切换下来后，被放在就绪队列（Ready Queue）中。等该进程下次被调度时，操作系统需要从上次中断的地方继续运行。

那么操作系统如何知道上一次运行的状态呢？

这里就产生了上下文切换，它的作用就是当一个进程被切换下来的时候，保存进程当前的相关信息，当该进程下一次被调度时，就可以通过保存的信息恢复到之前的运行状态接着执行。正是由于有这种机制存在，操作系统就可以在单个CPU上实现多任务了。

![](https://zyblog-1255449766.cos.ap-beijing.myqcloud.com/fc43f51e-38e3-48c0-832d-290298e86b4e.png)



上面已经提了就绪队列，那么就介绍一下操作系统主要的几个队列：

- Job Queue：所有的进程进入系统时，都会加入到作业队列（Job Queue），该队列也保存了所有进程的PCB信息。
- Ready Queue：进程已经分配到主存中，仅缺少CPU资源，任务等待执行。新创建的进程一般都是放在这里。
- Device Queues：有些进程由于设备的I/O被阻塞，会被放在这里。


它们的调度关系如下，

![scheduler](https://zyblog-1255449766.cos.ap-beijing.myqcloud.com/2beed83f-31ab-45d4-912c-9f1b84fd9749.jpeg)

*Image From: tutorialspoint.com*



### 进程的上下文切换

**上下文切换的信息**

为了让进程能够恢复到上一次运行时的状态，操作系统需要保存进程的哪些上下文信息呢？

1. 用户级上下文：进程的运行数据、用户堆栈信息及共享存储区
2. 寄存器级上下文：各类寄存器，比如其中最重要的PC（Program Counter）、栈指针、处理器状态寄存器等
3. 系统级上下文：进程控制块（PCB）、内存管理信息（MMU）、内核栈；


**上下文交换的代价**

首先需要介绍一下**用户态**和**内核态**。CPU总是处于内核态或者用户态中的其中一种状态。简单理解就是用户态的指令都是权限比较低、比较安全的指令，内核态的指令就是权限高的指令。

当我们进行上下文交换时，我们的进程总是需要权限高的指令，所以每次调度时，需要进入内核态才处理。每次上下文交换都需要几十纳秒到数微妙的CPU时间。

进程在上下文切换的过程中，消耗最多的是下面几个方面：

1. 进程的上下文切换后，由于对应的缓存已经失效，相当于缓存需要重新刷新一遍。
2. TLB的刷新（MMU/Memory Management Unit），TLB主要的作用是虚拟内存地址映射物理内存地址。
3. 寄存器的保存和恢复。

对于操作系统来说，一次上下文交换的消耗还是大的。






### 线程的上下文切换


若进程中只包含一个线程，那么对这个线程进行上下文切换也就是对于这个进程进行上下文切换，所带来的开销与上述进程切换是一样的。

如果一个进程有多个线程，这些线程共享着进程的虚拟内存空间，在这些线程之间进行上下文切换时，就不需要对TLB进行刷新，只需要保存线程各自的数据：寄存器、私有变量等。

另外，处于用户态线程的上下文切换，也不需要陷入系统内核即可进行上下文切换。



**放上StackOverflow的进程和线程上下文切换的解答：**


> 参考：[Thread context switch Vs. process context switch](https://stackoverflow.com/questions/5440128/thread-context-switch-vs-process-context-switch)
> 
> The main distinction between a thread switch and a process switch is that during a thread switch, the virtual memory space remains the same, while it does not during a process switch. Both types involve handing control over to the operating system kernel to perform the context switch. The process of switching in and out of the OS kernel along with the cost of switching out the registers is the largest fixed cost of performing a context switch.
> 
> A more fuzzy cost is that a context switch messes with the processors cacheing mechanisms. Basically, when you context switch, all of the memory addresses that the processor "remembers" in its cache effectively become useless. The one big distinction here is that when you change virtual memory spaces, the processor's Translation Lookaside Buffer (TLB) or equivalent gets flushed making memory accesses much more expensive for a while. This does not happen during a thread switch.


**说明**

> ⚠️ 后面不单独讨论进程的调度和线程的调度。引入操作系统级别的调度主要是对比即将介绍的Golang协程的调度。


<hr />

### 参考文献

- https://www.slideshare.net/matthewrdale/demystifying-the-go-scheduler
- Daniel Morsing - http://morsmachine.dk/go-scheduler
- 中文版：[Goroutine是如何工作的](https://cloud.tencent.com/developer/article/1065913)
- [Go语言调度器](https://studygolang.com/articles/6070)
- https://work-jlsun.github.io/2014/09/24/goroutine-scheduler.html
- [The Go netpoller](http://morsmachine.dk/netpoller)
- [Operating System - Processes](https://www.tutorialspoint.com/operating_system/os_processes.htm)
- [Operating System - Thread @geeksforgeeks.org](https://www.geeksforgeeks.org/operarting-system-thread/)
- [Are threads implemented as processes on Linux?](https://unix.stackexchange.com/questions/364660/are-threads-implemented-as-processes-on-linux)
- [Does linux schedule a process or a thread?](https://stackoverflow.com/questions/15601155/does-linux-schedule-a-process-or-a-thread)
- [OS process scheduling](https://www.tutorialspoint.com/operating_system/os_process_scheduling.htm)
- [操作系统（二）：进程与线程](http://blog.forec.cn/2016/11/22/os-concepts-2/)
- [CPU上下文切换]([http://wanggaoliang.club/2018/11/30/cpu%E4%B8%8A%E4%B8%8B%E6%96%87%E5%88%87%E6%8D%A2/](http://wanggaoliang.club/2018/11/30/cpu上下文切换/))


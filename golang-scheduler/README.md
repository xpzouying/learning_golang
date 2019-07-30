从这篇文章开始，通过几篇文章介绍Golang调度的细节。



来源：[个人博客](http://zouying.is/20190725-scheduler1-os/)



### 大纲

- 一、操作系统的调度
  - 进程/线程
  - 调度器/调度
  - 上下文交换（Context Switch）
- 二、Golang的调度
  - Golang调度是什么新奇东西？
  - Golang调度的优点
  - Goroutine是如何组成的
  - Goroutine内部是如何工作的
- 三、Demo
- 四、Golang调度的源码解析



<hr />
### 操作系统的调度

在介绍Golang的调度之前，先复习操作系统的调度。这也有利于理解Golang的调度，并与其进行对比。



**进程、线程**

首先介绍一下进程（Process）、线程（Thread）的概念。

进程就是一个运行状态的程序（Program）。进程简单可以理解为一系列待执行的任务/指令。

线程是进程中的一系列执行命令（a path of execution within a process），一个进程可以有多个线程。



**调度器是什么？我们为什么需要调度器？**

操作系统同时有很多任务要运行，任务只有分配到相应的硬件资源后（比如分配到CPU、内存）才能运行，然而可用的硬件资源（CPU、内存等）只有一份，如何给这些任务合理地分配硬件资源，让每个任务都得到合理地运行，这就需要调度器按照一定的策略对任务进行调度。



线程是操作系统运行的最基本单位，也是调度的基本单位。我们在本文中是讨论调度相关，就不区分进程和线程，进程和线程都为同一个概念。



> 早期版本中，Linus Torvalds在1996年提及过，线程和进程都是执行的上下文内容（“both threads and processes are treated as a ‘context of execution’”），其中包括的资源有：CPU状态、MMU状态、permissions、各种communication states（比如open files、signal handlers等等）



**什么是调度？**

调度就是将正在运行的进程从CPU上移除，然后从就绪队列中选择一个进程分配CPU资源让其运行。



上面提了就绪队列，那么主要的几个队列分别是：

- Job Queue：所有的进程进入系统时，都会加入到**作业队列（Job Queue）**，该队列也保存了所有进程的PCB信息。
- Ready Queue：进程已经分配到主存中，仅缺少CPU资源，任务等待执行。新创建的进程一般都是放在这里。
- Device Queues：有些进程由于设备的I/O被阻塞，会被放在这里。



![scheduler](https://zyblog-1255449766.cos.ap-beijing.myqcloud.com/2beed83f-31ab-45d4-912c-9f1b84fd9749)

*Image From: tutorialspoint.com*




**Context Switch / 上下文切换**

当进程运行过程中，由于某些原因（如I/O、CPU时间片用完等等），从CPU被切换下来后，被放在Ready Queue中。等该进程下次被调度时，操作系统需要从上次中断的地方接着执行。那么操作系统如何知道上一次运行的状态呢？



这里就用到了上下文切换，它的作用就是当一个进程被切换下来的时候，保存进程当前的相关信息，当该进程下一次被调度时，就可以通过保存的信息恢复到之前的运行状态接着执行。正是由于有这种机制存在，操作系统就可以在单个CPU上实现多任务了。



![](https://zyblog-1255449766.cos.ap-beijing.myqcloud.com/fc43f51e-38e3-48c0-832d-290298e86b4e)





**上下文交换的信息**

为了让进程能够恢复到上一次运行时的状态，操作系统需要保存进程的哪些上下文信息呢？



1. 用户级上下文：进程的运行数据、用户堆栈信息及共享存储区
2. 寄存器级上下文：各类寄存器，比如其中最重要的PC（program counter）、栈指针、处理器状态寄存器等
3. 系统级上下文：进程控制块（PCB）、内存管理信息（MMU）、内核栈；



**上下文交换的代价**



首先需要介绍一下**用户态**和**内核态**。CPU总是处于内核态或者用户态中的其中一种状态。简单理解就是用户态的指令都是权限比较低、比较安全的指令，内核态的指令就是权限高的指令。



当我们进行上下文交换时，我们的进程总是需要权限高的指令，所以每次调度时，需要进入内核态才处理。每次上下文交换都需要几十纳秒到数微妙的CPU时间。



上下文交换还包括MMU内存映射硬件的相关数据的保存和恢复。



所以可以看到对于操作系统来说，一次上下文交换的代价还是比较昂贵的。



>  ⚠️ 这里不单独讨论进程的调度和线程的调度。引入操作系统级别的调度主要是对比即将介绍的Golang协程的调度。





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


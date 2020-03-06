package main

// #include <stdint.h>
// #include <stdlib.h>
// int working(size_t begin, size_t end) {
// 	size_t res = 0;
// 	for (size_t i = begin; i < end; ++i) {
// 		if ((i%2) == 0) {
// 			res += i*2;
// 		} else {
// 			res -= i*2;
// 		}
// 	}
// 	return res;
// }
import "C"

import (
	"log"
	"os"
	"runtime"
	"runtime/trace"
	"sync"
)

const count = 9900000000

func working(begin, end int) int {
	res := 0
	for i := begin; i < end; i++ {
		if i%2 == 0 {
			res += i * 2
		} else {
			res -= i * 2
		}
	}

	return res
}

type req struct {
	begin int
	end   int
	resCh chan int
}

type worker struct {
	id    int
	reqCh chan *req
}

func (w *worker) DoGo() {
	for e := range w.reqCh {
		res := working(e.begin, e.end)

		e.resCh <- res
		close(e.resCh)
	}
}

func (w *worker) DoCgo() {
	for e := range w.reqCh {
		res := C.working(C.size_t(e.begin), C.size_t(e.end))

		e.resCh <- int(res)
		close(e.resCh)
	}
}

func main() {
	// trace profile
	f, err := os.Create("trace.out")
	if err != nil {
		log.Fatalf("failed to create trace output: %v", err)
	}
	defer f.Close()

	if err := trace.Start(f); err != nil {
		log.Fatalf("failed to start trace: %v", err)
	}
	defer trace.Stop()

	// start worker, the number of worker == cpu cores
	cores := runtime.NumCPU()
	reqCh := make(chan *req, 16)
	for i := 0; i < cores; i++ {
		w := worker{id: i, reqCh: reqCh}
		go w.DoGo() // Use GO
		// go w.DoCgo()  // Use cgo
	}

	// send request to worker
	reqNum := 10
	wg := new(sync.WaitGroup)
	wg.Add(reqNum)
	for i := 0; i < reqNum; i++ {
		i := i
		go func() {
			resCh := make(chan int)
			req := &req{begin: i, end: count, resCh: resCh}
			reqCh <- req // send request to worker for working

			res := <-resCh
			log.Printf("got res=%d", res)

			wg.Done()
		}()
	}

	wg.Wait()
}

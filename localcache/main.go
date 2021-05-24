package main

import (
	"log"
	"sync"
	"sync/atomic"
)

type cacheValue string

func getCache() func() cacheValue {
	var (
		cache atomic.Value
		mu    sync.Mutex
	)

	return func() cacheValue {

		if v, ok := cache.Load().(string); ok {
			return cacheValue(v)
		}

		mu.Lock()
		defer mu.Unlock()

		// 避免第一次，从cache.Load()中都没有获取到值后，
		// 进入mu.Lock()锁后的情况。
		// 如果不再Load一次的话，那么就会出现多次Store的情况。
		if v, ok := cache.Load().(string); ok {
			return cacheValue(v)
		}

		s := cacheValue("hello world")
		log.Printf("cache.Store: %s", s)
		cache.Store(s) // only store once

		return s
	}
}

func main() {

	fn := getCache()

	cacheValue := fn()

	wg := new(sync.WaitGroup)
	wg.Add(20)
	for i := 0; i < 20; i++ {
		go func() {

			log.Println("cacheValue: ", cacheValue)
			wg.Done()
		}()
	}

	wg.Wait()
}

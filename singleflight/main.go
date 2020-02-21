package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/golang/groupcache/singleflight"
)

func doReq() (interface{}, error) {
	log.Printf("do http request")

	c := http.Client{Timeout: 3 * time.Second}

	resp, err := c.Get("https://www.163.com/")
	if err != nil {
		log.Printf("http request error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("http status != HTTP.OK (200)")
		return nil, fmt.Errorf("http status code = %d", resp.StatusCode)
	}

	return "finish", nil
}

func main() {
	var g singleflight.Group
	var wg sync.WaitGroup

	count := 5

	wg.Add(count)

	for i := 0; i < count; i++ {
		go func(i int) {
			defer wg.Done()

			res, err := g.Do("163", doReq)
			if err != nil {
				log.Printf("goroutine-%d do error: %v", i, err)
				return
			}

			log.Printf("goroutine-%d, result: %s", i, res)
		}(i)
	}

	wg.Wait()
}

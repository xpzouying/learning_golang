package main

import (
	"bytes"
	"net/http"
	"strconv"
	"sync"

	"github.com/rs/zerolog"

	"github.com/sirupsen/logrus"
)

var counter = map[string]int{}
var mu sync.Mutex // mutex for counter

var pool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
		// buf := bytes.NewBuffer(make([]byte, 0, 8192))
		// return buf
	},
}

var logger zerolog.Logger

func handleHello(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	mu.Lock()
	counter[name]++
	cnt := []byte(strconv.Itoa(counter[name]))
	mu.Unlock()

	buf := pool.Get().(*bytes.Buffer)
	buf.Reset()
	buf.Write([]byte("<h1 style='color: "))
	buf.Write([]byte(r.FormValue("color")))
	buf.Write([]byte("'>Welcome!</h1> <p>Name: "))
	buf.Write([]byte(name))
	buf.Write([]byte("</p> <p>Count: "))
	buf.Write(cnt)
	w.Write(buf.Bytes())
	pool.Put(buf)

	logbuf := pool.Get().(*bytes.Buffer)
	logbuf.Reset()
	logbuf.Write([]byte("visited name="))
	logbuf.Write([]byte(name))
	logbuf.Write([]byte("count="))
	logbuf.Write(cnt)
	logger.Info().Msg(logbuf.String())
	pool.Put(logbuf)
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	http.HandleFunc("/hello", handleHello)
	logrus.Fatal(http.ListenAndServe(":8080", nil))
}

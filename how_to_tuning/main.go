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
	},
}

var logger zerolog.Logger

func handleHello(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	mu.Lock()
	defer mu.Unlock()
	counter[name]++

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	buf := pool.Get().(*bytes.Buffer)
	buf.Reset()
	buf.Write([]byte("<h1 style='color: "))
	buf.Write([]byte(r.FormValue("color")))
	buf.Write([]byte("'>Welcome!</h1> <p>Name: "))
	buf.Write([]byte(name))
	buf.Write([]byte("</p> <p>Count: "))
	b := strconv.AppendInt(buf.Bytes(), int64(counter[name]), 10)
	b = append(b, []byte("</p>")...)
	w.Write(b)
	pool.Put(buf)

	logbuf := pool.Get().(*bytes.Buffer)
	logbuf.Reset()
	logbuf.Write([]byte("visited name="))
	logbuf.Write([]byte(name))
	logbuf.Write([]byte("count="))
	strconv.AppendInt(logbuf.Bytes(), int64(counter[name]), 10)
	logger.Info().Msg(logbuf.String())
	pool.Put(logbuf)
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	http.HandleFunc("/hello", handleHello)
	logrus.Fatal(http.ListenAndServe(":8080", nil))
}

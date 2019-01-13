package main

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/sirupsen/logrus"
)

var counter = map[string]int{}
var mu sync.Mutex // mutex for counter

var pool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func handleHello(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	mu.Lock()
	defer mu.Unlock()
	counter[name]++

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	// w.Write([]byte("<h1 style='color: " + r.FormValue("color") +
	// 	"'>Welcome!</h1> <p>Name: " + name + "</p> <p>Count: " + fmt.Sprint(counter[name]) + "</p>"))

	// use - fmt.Fprintf
	// fmt.Fprintf(w, "<h1 style='color: %s>Welcome!</h1> <p>Name: %s</p> <p>Count: %d</p>",
	// 	r.FormValue("color"),
	// 	name,
	// 	counter[name],
	// )

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
	logbuf.WriteString(fmt.Sprintf("visited name=%s count=%d", name, counter[name]))
	logrus.Infof("%s", logbuf.String())
	pool.Put(logbuf)

	// logrus.WithFields(logrus.Fields{
	// 	"module": "main",
	// 	"name":   name,
	// 	"count":  counter[name],
	// }).Infof("visited")
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})

	http.HandleFunc("/hello", handleHello)
	logrus.Fatal(http.ListenAndServe(":8080", nil))
}

package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
)

func BenchmarkHandleFunc(b *testing.B) {
	b.ReportAllocs()

	logrus.SetOutput(ioutil.Discard)

	rw := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/hello?name=zouying", nil)

	for i := 0; i < b.N; i++ {
		handleHello(rw, req)
	}
}

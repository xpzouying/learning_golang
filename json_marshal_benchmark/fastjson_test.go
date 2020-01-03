package main

import (
	"encoding/json"
	"testing"

	"github.com/valyala/fastjson"
)

func BenchmarkImageBase64_fastjson(b *testing.B) {
	b.ReportAllocs()

	s := Base64ImgStruct{Img: Image160KB}
	data, _ := json.Marshal(s)

	b.ResetTimer()
	var p fastjson.Parser
	for i := 0; i < b.N; i++ {
		v, err := p.Parse(string(data))
		if err != nil {
			b.Error(err)
		}
		v.GetStringBytes("img")
	}
}

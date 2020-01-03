package main

import (
	"encoding/json"
	"testing"
)

func BenchmarkImageBase64_stdjson(b *testing.B) {
	b.ReportAllocs()

	s := Base64ImgStruct{Img: Image160KB}
	data, _ := json.Marshal(s)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := new(Base64ImgStruct)
		if err := json.Unmarshal(data, s); err != nil {
			b.Error(err)
		}
	}
}

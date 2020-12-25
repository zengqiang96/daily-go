package str

import (
	"testing"
	"unsafe"
)

var (
	byteData = []byte("一去二三里，烟村四五家")
)

func BenchmarkSlice2String(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = string(byteData)
	}
}

func BenchmarkSlice2String2(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SliceByteToString(byteData)
	}
}

func SliceByteToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

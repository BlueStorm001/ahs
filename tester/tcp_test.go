package tester

import (
	"sync"
	"testing"
)

var pool sync.Map

type tests struct {
	n int
}

func init() {
	for i := 0; i < 1000; i++ {
		pool.Store(i, &tests{n: i})
	}
}

func Benchmark_Tester(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			//pool.Range(func(key, value interface{}) bool {
			//	return true
			//})

		}
	})
}

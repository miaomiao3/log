package benchmarks

import (
	"testing"
	"github.com/miaomiao3/log"
)

func BenchmarkDefaultLogger(b *testing.B) {
	b.Logf("default logger test")
	b.Run("miaomiao3", func(b *testing.B) {
		b.ResetTimer()
		b.SetBytes(1000)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				log.Debug("123")
			}
		})
	})
}
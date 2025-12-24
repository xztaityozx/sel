package parser

import "testing"

func BenchmarkParse_SingleIndex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Parse([]string{"1"})
	}
}

func BenchmarkParse_MultipleIndex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Parse([]string{"1", "2", "3", "4", "5"})
	}
}

func BenchmarkParse_Range(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Parse([]string{"1:10"})
	}
}

func BenchmarkParse_RangeWithStep(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Parse([]string{"1:10:2"})
	}
}

func BenchmarkParse_Regexp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Parse([]string{"/start/:/end/"})
	}
}

func BenchmarkParse_Complex(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = Parse([]string{"1", "2:10", "/start/:/end/", "-1", "3::2"})
	}
}

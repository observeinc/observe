package main

import "testing"

var Q = make([]int, 100000)

func FuncA() {
	var m int
	for i := 0; i != len(Q); i++ {
		m += Q[i]
	}
	if m != 0 {
		panic("boo")
	}
}

func FuncB() {
	var m int
	for i, n := 0, len(Q); i != n; i++ {
		m += Q[i]
	}
	if m != 0 {
		panic("boo")
	}
}

func BenchmarkA(b *testing.B) {
	for n := 0; n != b.N; n++ {
		FuncA()
	}
}

func BenchmarkB(b *testing.B) {
	for n := 0; n != b.N; n++ {
		FuncB()
	}
}

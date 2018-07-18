package main

import (
	"testing"
)

func TestSimple(t *testing.T) {
	got := Filename("/logs/2017/12/21/04/logs_1488781d-3d24-4ac4-a3cb-3b4dc82dd3a0.txt")
	want := "logs_1488781d-3d24-4ac4-a3cb-3b4dc82dd3a0.txt"
	if got != want {
		t.Fatalf("want %v, but %v:", want, got)
	}
}

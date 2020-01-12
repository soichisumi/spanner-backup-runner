package main

import (
	"github.com/soichisumi/go-util/slice"
)

const (
	from = 1
	to   = 10
)

func main() {
	seq := slice.Sequence(from, to-from, 1)
	for _, v := range seq {

	}
}

package main

import (
	"github.com/mindstand/gogm/cmd/gogm/gen"
	"log"
)

func main() {
	directory := "/home/erictg97/mindstand/repos/gogm/testing_"

	err := gen.Generate(directory)
	if err != nil {
		log.Fatal(err)
	}
}


package main

import (
	"fmt"
	"os"

	md5 "github.com/SergeyCherepiuk/md5all/pkg"
)

func main() {
	results := md5.Sum(os.Args[1])
	for result := range results {
		if result.Err != nil {
			fmt.Printf("%s -> error: %s\n", result.Path, result.Err.Error())
		} else {
			fmt.Printf("%s -> %x\n", result.Path, result.Sum)
		}
	}
}

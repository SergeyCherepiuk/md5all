package main

import (
	"crypto/md5"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var N_WORKERS = runtime.GOMAXPROCS(0)

type result struct {
	path string
	sum  [md5.Size]byte
	err  error
}

func generatePaths(root string) <-chan string {
	paths := make(chan string, N_WORKERS)
	go func() {
		if err := filepath.Walk(os.Args[1], func(path string, info fs.FileInfo, err error) error {
			if err == nil && info.Mode().IsRegular() {
				paths <- path
			}
			return err
		}); err != nil {
			log.Fatal(err)
		}
		close(paths)
	}()
	return paths
}

func checksum(paths <-chan string) <-chan result {
	results := make(chan result)
	go func() {
		for path := range paths {
			bytes, err := os.ReadFile(path)
			results <- result{path, md5.Sum(bytes), err}
		}
		close(results)
	}()
	return results
}

func fanout(paths <-chan string) []<-chan result {
	resultChannels := make([]<-chan result, N_WORKERS)
	for i := 0; i < N_WORKERS; i++ {
		resultChannels[i] = checksum(paths)
	}
	return resultChannels
}

func fanin(resultChannels []<-chan result) <-chan result {
	wg := sync.WaitGroup{}
	wg.Add(len(resultChannels))

	results := make(chan result, N_WORKERS)
	push := func(ch <-chan result) {
		for r := range ch {
			results <- r
		}
		wg.Done()
	}

	go func() {
		for _, ch := range resultChannels {
			go push(ch)
		}
		wg.Wait()
		close(results)
	}()

	return results
}

func main() {
	paths := generatePaths(os.Args[1])
	resultChannels := fanout(paths)
	results := fanin(resultChannels)

	for result := range results {
		if result.err == nil {
			fmt.Printf("%s -> %x\n", result.path, result.sum)
		} else {
			fmt.Printf("%s -> error: %s\n", result.path, result.err.Error())
		}
	}
}

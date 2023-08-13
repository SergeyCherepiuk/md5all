package md5

import (
	"crypto/md5"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var N_WORKERS = runtime.GOMAXPROCS(0)

type Result struct {
	Path string
	Sum  [md5.Size]byte
	Err  error
}

func getPaths(root string) <-chan string {
	paths := make(chan string, N_WORKERS)
	go func() {
		defer close(paths)
		if err := filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
			if err == nil && info.Mode().IsRegular() {
				paths <- path
			}
			return err
		}); err != nil {
			log.Fatal(err)
		}
	}()
	return paths
}

func checksum(paths <-chan string) <-chan Result {
	results := make(chan Result)
	go func() {
		defer close(results)
		for path := range paths {
			bytes, err := os.ReadFile(path)
			results <- Result{path, md5.Sum(bytes), err}
		}
	}()
	return results
}

func fanout(paths <-chan string) []<-chan Result {
	resultChannels := make([]<-chan Result, N_WORKERS)
	for i := 0; i < N_WORKERS; i++ {
		resultChannels[i] = checksum(paths)
	}
	return resultChannels
}

func fanin(resultChannels []<-chan Result) <-chan Result {
	wg := sync.WaitGroup{}
	wg.Add(len(resultChannels))

	results := make(chan Result, N_WORKERS)
	push := func(ch <-chan Result) {
		for r := range ch {
			results <- r
		}
		wg.Done()
	}

	go func() {
		defer close(results)
		for _, ch := range resultChannels {
			go push(ch)
		}
		wg.Wait()
	}()

	return results
}

func MD5All(path string) <-chan Result {
	paths := getPaths(path)
	resultChannels := fanout(paths)
	return fanin(resultChannels)
}
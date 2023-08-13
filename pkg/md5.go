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

var n_workers = runtime.GOMAXPROCS(0)

type Result struct {
	Path string
	Sum  [md5.Size]byte
	Err  error
}

func getPaths(root string) <-chan string {
	paths := make(chan string, n_workers)
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
	workerChs := make([]<-chan Result, n_workers)
	for i := 0; i < n_workers; i++ {
		workerChs[i] = checksum(paths)
	}
	return workerChs
}

func fanin(workerChs ...<-chan Result) <-chan Result {
	wg := sync.WaitGroup{}
	wg.Add(len(workerChs))

	results := make(chan Result, n_workers)
	push := func(ch <-chan Result) {
		for r := range ch {
			results <- r
		}
		wg.Done()
	}

	go func() {
		defer close(results)
		for _, ch := range workerChs {
			go push(ch)
		}
		wg.Wait()
	}()

	return results
}

func Sum(root string) <-chan Result {
	paths := getPaths(root)
	workerChs := fanout(paths)
	return fanin(workerChs...)
}

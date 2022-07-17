package main

import (
	"fmt"
	"mgrep/worker"
	"mgrep/worklist"
	"os"
	"path/filepath"
	"sync"

	"github.com/alexflint/go-arg"
)

//where all go routines and waitgroups are implemented

func discoverDirs(wl *worklist.Worklist, path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		fmt.Println("Readdir error:", err)
		return
	}
	for _, entry := range entries {
		nextPath := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			discoverDirs(wl, nextPath)
		} else {
			wl.Add(worklist.NewJob(nextPath))
		}
	}
}

var args struct {
	SearchTerm string `arg:"positional,required"`
	SearchDir  string `arg:"positional,required"`
}

func main() {
	arg.MustParse(&args) //ensures arguments are valid

	var workersWg sync.WaitGroup
	wl := worklist.New(100) //100 jobs before adding to channel is blocked

	results := make(chan worker.Result, 100)

	numWorkers := 10

	workersWg.Add(1)

	go func() {
		defer workersWg.Done()
		discoverDirs(&wl, args.SearchDir)
		wl.Finalize(numWorkers)
	}()

	for i := 0; i < numWorkers; i++ {
		workersWg.Add(1)
		go func() {
			defer workersWg.Done()
			for {
				workEntry := wl.Next()
				if workEntry.Path != "" {
					workerResult := worker.FindInFile(workEntry.Path, args.SearchTerm)
					if workerResult != nil {
						for _, r := range workerResult.Inner {
							results <- r
						}
					}
				} else {
					return
				}
			}
		}()
	}

	//wait on workers to finish and display results
	blockWorkersWg := make(chan struct{}) //will select on this channel
	go func() {
		workersWg.Wait()
		close(blockWorkersWg)
	}()

	var displayWg sync.WaitGroup

	displayWg.Add(1)
	go func() {
		for {
			select {
			case r := <-results:
				fmt.Printf("%v[%v]: %v\n", r.Path, r.LineNum, r.Line)
			case <-blockWorkersWg:
				if len(results) == 0 {
					displayWg.Done()
					return
				}
			}
		}
	}()
	displayWg.Wait()
}

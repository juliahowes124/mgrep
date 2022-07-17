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

//traverse directory and add files to worklist
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
	arg.MustParse(&args) //ensure arguments are valid

	numWorkers := 10
	var workersWg sync.WaitGroup
	wl := worklist.New(100) //100 jobs before channel is blocked

	results := make(chan worker.Result, 100)

	workersWg.Add(1)

	//Build waitlist of files to process
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
				workEntry := wl.Next() //grab next file in the worklist
				if workEntry.Path != "" {
					workerResults := worker.FindInFile(workEntry.Path, args.SearchTerm)
					if workerResults != nil {
						for _, r := range workerResults.Inner {
							results <- r
						}
					}
				} else {
					return
				}
			}
		}()
	}

	blockWorkersWg := make(chan struct{})
	//since Wait() is blocking, call in a go routine so program can print results while waiting for jobs to finish
	go func() {
		workersWg.Wait()
		close(blockWorkersWg)
	}()

	var displayWg sync.WaitGroup

	//wait on workers to finish and display results
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

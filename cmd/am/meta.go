package main

import (
	"fmt"
	"log"
	"runtime"
	"sync"

	"mcquay.me/arrange"
)

func meta(files []string) {
	workers := runtime.NumCPU()
	if *cores != 0 {
		workers = *cores
	}
	fc := make(chan arrange.Media)

	go func() {
		wg := &sync.WaitGroup{}
		s := make(chan bool, workers)
		for _, f := range files {
			wg.Add(1)
			go func(pth string) {
				s <- true
				pf, err := arrange.ParseFile(pth)
				if err != nil {
					log.Printf("%+v", err)
				}

				fc <- pf
				<-s
				wg.Done()
			}(f)
		}
		wg.Wait()
		close(fc)
	}()
	for f := range fc {
		fmt.Printf("%+v: %v\n", f.Time, f.Path)
	}
}

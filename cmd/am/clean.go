package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"mcquay.me/arrange"
)

func clean(dir string) error {
	dateDir := filepath.Join(dir, "date")
	if _, err := os.Stat(dateDir); os.IsNotExist(err) {
		return fmt.Errorf("couldn't find 'date' dir in %q", dir)
	}

	work := arrange.Source(dateDir)
	streams := []<-chan arrange.Media{}
	errs := []<-chan error{}

	workers := runtime.NumCPU()
	if *cores != 0 {
		workers = *cores
	}

	for w := 0; w < workers; w++ {
		s, e := arrange.MissingLink(arrange.Parse(work), dir)
		streams = append(streams, s)
		errs = append(errs, e)
	}

	var err error
	go func() {
		for e := range eMerge(errs) {
			log.Printf("%+v", e)
			err = fmt.Errorf("%v, %v", err, e)
		}
	}()

	for m := range arrange.Merge(streams) {
		log.Printf("%q > %q", m.Path, m.Content(dir))
		if err := os.Remove(m.Path); err != nil {
			log.Printf("%+v", err)
		}
		if err := os.Link(m.Content(dir), m.Path); err != nil {
			log.Printf("%+v", err)
		}
	}

	return err
}

func eMerge(cs []<-chan error) <-chan error {
	out := make(chan error)
	var wg sync.WaitGroup
	output := func(c <-chan error) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}
	for _, c := range cs {
		go output(c)
	}
	wg.Add(len(cs))
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

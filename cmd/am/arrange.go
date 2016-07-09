package main

import (
	"fmt"
	"log"
	"runtime"

	"mcquay.me/arrange"
)

func arr(indir, outdir string) error {
	if err := arrange.PrepOutput(outdir); err != nil {
		return fmt.Errorf("problem creating directory structure: %v", err)
	}

	work := arrange.Source(indir)
	streams := []<-chan arrange.Media{}

	workers := runtime.NumCPU()
	if *cores != 0 {
		workers = *cores
	}

	for w := 0; w < workers; w++ {
		streams = append(streams, arrange.Parse(work))
	}

	st := stats{}
	for err := range arrange.Move(arrange.Merge(streams), outdir) {
		st.total++
		if err != nil {
			switch err.(type) {
			case arrange.Dup:
				st.dupes++
			default:
				log.Printf("%+v", err)
			}
		} else {
			st.moved++
		}
	}

	log.Printf("dupes: %+v", st.dupes)
	log.Printf("moved: %+v", st.moved)
	log.Printf("total: %+v", st.total)
	return nil
}

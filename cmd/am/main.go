package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"

	"mcquay.me/arrange"
)

const usage = "aj <indir> <outdir>"

type stats struct {
	total int
	dupes int
	moved int
}

var cores = flag.Int("cores", 0, "how many threads to use")

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile)
	if len(flag.Args()) != 2 {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
		os.Exit(1)
	}
	in, out := flag.Args()[0], flag.Args()[1]

	if err := arrange.PrepOutput(out); err != nil {
		fmt.Fprintf(os.Stderr, "problem creating directory structure: %v", err)
		os.Exit(1)
	}

	work := arrange.Source(in)
	streams := []<-chan arrange.Media{}

	workers := runtime.NumCPU()
	if *cores != 0 {
		workers = *cores
	}

	for w := 0; w < workers; w++ {
		streams = append(streams, arrange.Parse(work))
	}

	st := stats{}
	for err := range arrange.Move(arrange.Merge(streams), out) {
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
}

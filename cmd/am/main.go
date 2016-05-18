package main

import (
	"fmt"
	"log"
	"os"

	"mcquay.me/arrange"
)

const usage = "aj <indir> <outdir>"

type stats struct {
	total int
	dupes int
	moved int
}

func main() {
	log.SetFlags(log.Lshortfile)
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
		os.Exit(1)
	}
	in, out := os.Args[1], os.Args[2]

	if err := arrange.PrepOutput(out); err != nil {
		fmt.Fprintf(os.Stderr, "problem creating directory structure: %v", err)
		os.Exit(1)
	}

	exts := map[string]bool{
		// images
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,

		// videos
		".mov": true,
		".mp4": true,
		".m4v": true,
	}

	work := arrange.Source(in, exts)
	streams := []<-chan arrange.File{}

	for w := 0; w < 16; w++ {
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

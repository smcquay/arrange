package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

const usage = "am <arr|clean|meta> [flags]"
const arrUsage = "am arr [-h|-cores=N] <in> <out>"
const cleanUsage = "am clean [-h|-cores=N] <directory>"
const metaUsage = "am meta [-h|-cores=N] <file0> <file1> ... <fileN>"

type stats struct {
	total int
	dupes int
	moved int
}

var cores = flag.Int("cores", 0, "how many threads to use")

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
		os.Exit(1)
	}
	var sub string
	sub, os.Args = os.Args[1], os.Args[1:]

	flag.Parse()
	log.SetFlags(log.Lshortfile)

	switch sub {
	case "ar", "arr", "arrange":
		args := flag.Args()
		if len(args) != 2 {
			fmt.Fprintf(os.Stderr, "%s\n", arrUsage)
			os.Exit(1)
		}
		in, out := args[0], args[1]
		if err := arr(in, out); err != nil {
			fmt.Fprintf(os.Stderr, "problem arranging media: %v\n", err)
			os.Exit(1)
		}
	case "c", "cl", "clean":
		args := flag.Args()
		if len(args) != 1 {
			fmt.Fprintf(os.Stderr, "%s\n", cleanUsage)
			os.Exit(1)
		}
		dir := args[0]
		if err := clean(dir); err != nil {
			fmt.Fprintf(os.Stderr, "problem cleaning: %v\n", err)
			os.Exit(1)
		}
	case "m", "meta":
		args := flag.Args()
		if len(args) < 1 {
			fmt.Fprintf(os.Stderr, "%s\n", metaUsage)
			os.Exit(1)
		}
		meta(args)
	default:
		fmt.Fprintf(os.Stderr, "%s\n", usage)
		os.Exit(1)
	}

}

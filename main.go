package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/rwcarlsen/goexif/exif"
)

const usage = "picmv <indir> <outdir>"

type input struct {
	path  string
	year  string
	month string
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
		os.Exit(1)
	}
	in, out := os.Args[1], os.Args[2]
	log.Printf("%+v", in)
	log.Printf("%+v", out)

	count := 0
	files := make(chan input, 20)
	// seen := map[string]bool{}

	go func() {
		err := filepath.Walk(
			in,
			func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}
				f, err := os.Open(path)
				if err != nil {
					return fmt.Errorf("problem opening file: %v", err)
				}
				defer f.Close()
				x, err := exif.Decode(f)
				if err != nil {
					log.Printf("problem decoding exif data %q: %v", path, err)
					return nil
				}
				tm, err := x.DateTime()
				if err != nil {
					return fmt.Errorf("problem getting datetime from pic %v: %v", path, err)
				}
				files <- input{
					path:  path,
					year:  fmt.Sprintf("%04d", tm.Year()),
					month: fmt.Sprintf("%02d", tm.Month()),
				}
				return nil
			},
		)
		if err != nil {
			log.Printf("%+v", err)
		}
		close(files)
	}()

	for in := range files {
		log.Printf("%+v", in)
		count++
	}
	log.Printf("%+v", count)
}

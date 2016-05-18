package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

const usage = "aj <indir> <outdir>"

type file interface {
	move(root string) error
}

type jpg struct {
	path  string
	hash  string
	year  string
	month string
	time  string
}

func (j jpg) move(root string) error {
	f, err := os.Open(j.path)
	if err != nil {
		return fmt.Errorf("problem opening jpg file: %v", err)
	}
	defer f.Close()

	content := filepath.Join(root, "content", j.hash[:2], j.hash[2:]+".jpg")

	if _, err := os.Stat(content); !os.IsNotExist(err) {
		return dup{content}
	}

	out, err := os.Create(content)
	if err != nil {
		return fmt.Errorf("could not create output file: %v", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, f); err != nil {
		return fmt.Errorf("trouble copying file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "date", j.year, j.month), 0755); err != nil {
		return fmt.Errorf("problem creating date directory: %v", err)
	}

	date := filepath.Join(root, "date", j.year, j.month, j.time)
	name := date + ".jpg"
	for i := 0; i < 10000; i++ {
		if _, err := os.Stat(name); os.IsNotExist(err) {
			break
		}
		name = fmt.Sprintf("%s_%04d.jpg", date, i)
	}

	// TODO: or maybe symlinking? (issue #2)
	// rel := filepath.Join("..", "..", "..", "content", j.hash[:2], j.hash[2:]+".jpg")
	// return os.Symlink(rel, name)
	return os.Link(content, name)
}

type media struct {
	path string
}

func (m media) move(root string) error {
	return errors.New("NYI")
}

type stats struct {
	total int
	dupes int
	moved int
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
		os.Exit(1)
	}
	in, out := os.Args[1], os.Args[2]

	if err := prepOutput(out); err != nil {
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

	work := source(in, exts)
	streams := []<-chan file{}

	for w := 0; w < 16; w++ {
		streams = append(streams, parse(work))
	}

	st := stats{}
	for err := range move(merge(streams), out) {
		st.total++
		if err != nil {
			switch err.(type) {
			case dup:
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

func prepOutput(root string) error {
	for i := 0; i <= 0xff; i++ {
		dirname := filepath.Join(root, "content", fmt.Sprintf("%02x", i))
		if err := os.MkdirAll(dirname, 0755); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "date"), 0755); err != nil {
		return err
	}
	return nil
}

func source(root string, exts map[string]bool) <-chan string {
	out := make(chan string)
	go func() {
		err := filepath.Walk(
			root,
			func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}
				ext := strings.ToLower(filepath.Ext(path))
				if _, ok := exts[ext]; ok {
					out <- path
				} else {
					log.Printf("ignoring: %q", path)
				}
				return nil
			},
		)
		if err != nil {
			log.Printf("problem during crawl: %+v", err)
		}
		close(out)
	}()
	return out
}

func parse(in <-chan string) <-chan file {
	out := make(chan file)
	go func() {
		for path := range in {
			f, err := _parse(path)
			if err != nil {
				switch err.(type) {
				case notMedia:
					log.Printf("%+v", err)
				default:
					log.Printf("parse error: %+v", err)
				}
				continue
			} else {
				out <- f
			}
		}
		close(out)
	}()

	return out
}

func move(in <-chan file, root string) <-chan error {
	out := make(chan error)
	go func() {
		for i := range in {
			out <- i.move(root)
		}
		close(out)
	}()
	return out
}

func _parse(path string) (file, error) {
	ext := strings.ToLower(filepath.Ext(path))
	var r file
	switch ext {
	default:
		return nil, notMedia{path}
	case ".jpg", ".jpeg":
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("problem opening file: %v", err)
		}
		defer f.Close()

		if _, err := jpeg.DecodeConfig(f); err != nil {
			return nil, notMedia{path}
		}
		if _, err := f.Seek(0, 0); err != nil {
			return nil, fmt.Errorf("couldn't seek back in file: %v", err)
		}

		// try a few things for a time value
		var t time.Time
		{
			success := false
			if t, err = parseExif(f); err == nil {
				success = true
			}
			if !success {
				log.Printf("no exif for %q: %+v", path, err)
				t, err = mtime(path)
			}
			if err != nil {
				return nil, fmt.Errorf("unable to calculate reasonble time for jpg %q: %v", path, err)
			}
		}

		if _, err := f.Seek(0, 0); err != nil {
			return nil, fmt.Errorf("couldn't seek back in file: %v", err)
		}
		hash := md5.New()
		if _, err := io.Copy(hash, f); err != nil {
			return nil, fmt.Errorf("problem calculating checksum on %q: %v", path, err)
		}
		r = jpg{
			path:  path,
			hash:  fmt.Sprintf("%x", hash.Sum(nil)),
			year:  fmt.Sprintf("%04d", t.Year()),
			month: fmt.Sprintf("%02d", t.Month()),
			time:  fmt.Sprintf("%d", t.UnixNano()),
		}
	case ".png":
		return nil, fmt.Errorf("NYI: %q", path)
	case ".mov", ".mp4", ".m4v":
		return nil, fmt.Errorf("NYI: %q", path)
	}
	return r, nil
}

func merge(cs []<-chan file) <-chan file {
	out := make(chan file)
	var wg sync.WaitGroup
	output := func(c <-chan file) {
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

type notMedia struct {
	path string
}

func (nm notMedia) Error() string {
	return fmt.Sprintf("not media: %q", nm.path)
}

type dup struct {
	path string
}

func (d dup) Error() string {
	return fmt.Sprintf("dup: %q", d.path)
}

func parseExif(f io.Reader) (time.Time, error) {
	ti := time.Time{}
	x, err := exif.Decode(f)
	if err != nil {
		if exif.IsCriticalError(err) {
			return ti, err
		}
	}
	tm, err := x.DateTime()
	if err != nil {
		return ti, fmt.Errorf("no datetime in an ostensibly valid exif %v", err)
	}
	return tm, nil
}

func mtime(path string) (time.Time, error) {
	ti := time.Time{}
	s, err := os.Stat(path)
	if err != nil {
		return ti, fmt.Errorf("failure to collect times from stat: %v", err)
	}
	return s.ModTime(), nil
}

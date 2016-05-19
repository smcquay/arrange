package arrange

import (
	"crypto/md5"
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var exts map[string]bool

func init() {
	exts = map[string]bool{
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
}

type File interface {
	Move(root string) error
}

func PrepOutput(root string) error {
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

func Source(root string) <-chan string {
	out := make(chan string)
	go func() {
		err := filepath.Walk(
			root,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					return nil
				}
				ext := strings.ToLower(filepath.Ext(path))
				if _, ok := exts[ext]; ok {
					out <- path
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

func Parse(in <-chan string) <-chan File {
	out := make(chan File)
	go func() {
		for path := range in {
			f, err := _parse(path)
			if err != nil {
				switch err.(type) {
				case NotMedia:
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

func Move(in <-chan File, root string) <-chan error {
	out := make(chan error)
	go func() {
		for i := range in {
			out <- i.Move(root)
		}
		close(out)
	}()
	return out
}

func _parse(path string) (File, error) {
	ext := strings.ToLower(filepath.Ext(path))
	var r File
	switch ext {
	default:
		return nil, NotMedia{path}
	case ".jpg", ".jpeg":
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("problem opening file: %v", err)
		}
		defer f.Close()

		if _, err := jpeg.DecodeConfig(f); err != nil {
			return nil, NotMedia{path}
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
		r = Image{
			Path: path,
			Hash: fmt.Sprintf("%x", hash.Sum(nil)),
			Time: t,
		}
	case ".png":
		return nil, fmt.Errorf("NYI: %q", path)
	case ".mov", ".mp4", ".m4v":
		return nil, fmt.Errorf("NYI: %q", path)
	}
	return r, nil
}

func Merge(cs []<-chan File) <-chan File {
	out := make(chan File)
	var wg sync.WaitGroup
	output := func(c <-chan File) {
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

package arrange

import (
	"crypto/md5"
	"fmt"
	"image/gif"
	"image/jpeg"
	"image/png"
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
		".avi": true,
	}
}

func mtime(path string) (time.Time, error) {
	ti := time.Time{}
	s, err := os.Stat(path)
	if err != nil {
		return ti, fmt.Errorf("failure to collect times from stat: %v", err)
	}
	return s.ModTime(), nil
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

func Parse(in <-chan string) <-chan Media {
	out := make(chan Media)
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

func Move(in <-chan Media, root string) <-chan error {
	out := make(chan error)
	go func() {
		for i := range in {
			out <- i.Move(root)
		}
		close(out)
	}()
	return out
}

func _parse(path string) (Media, error) {
	ext := strings.ToLower(filepath.Ext(path))
	var r Media
	hash := md5.New()
	var t time.Time

	f, err := os.Open(path)
	if err != nil {
		return r, fmt.Errorf("problem opening file: %v", err)
	}
	defer f.Close()

	switch ext {
	default:
		return r, NotMedia{path}
	case ".jpg", ".jpeg":
		if _, err := jpeg.DecodeConfig(f); err != nil {
			return r, NotMedia{path}
		}
		if _, err := f.Seek(0, 0); err != nil {
			return r, fmt.Errorf("couldn't seek back in file: %v", err)
		}

		// try a few things for a time value
		{
			success := false
			if t, err = parseExif(f); err == nil {
				success = true
			}
			if !success {
				t, err = mtime(path)
			}
			if err != nil {
				return r, fmt.Errorf("unable to calculate reasonble time for jpg %q: %v", path, err)
			}
		}
	case ".png":
		if _, err := png.DecodeConfig(f); err != nil {
			return r, NotMedia{path}
		}
		if _, err := f.Seek(0, 0); err != nil {
			return r, fmt.Errorf("couldn't seek back in file: %v", err)
		}

		t, err = mtime(path)
		if err != nil {
			return r, fmt.Errorf("unable to calculate reasonble time for media %q: %v", path, err)
		}
	case ".gif":
		if _, err := gif.DecodeConfig(f); err != nil {
			return r, NotMedia{path}
		}
		if _, err := f.Seek(0, 0); err != nil {
			return r, fmt.Errorf("couldn't seek back in file: %v", err)
		}

		t, err = mtime(path)
		if err != nil {
			return r, fmt.Errorf("unable to calculate reasonble time for media %q: %v", path, err)
		}
	case ".mov", ".mp4", ".m4v", ".avi":
		t, err = mtime(path)
		if err != nil {
			return r, fmt.Errorf("unable to calculate reasonble time for media %q: %v", path, err)
		}
	}

	if _, err := f.Seek(0, 0); err != nil {
		return r, fmt.Errorf("couldn't seek back in file: %v", err)
	}
	if _, err := io.Copy(hash, f); err != nil {
		return r, fmt.Errorf("problem calculating checksum on %q: %v", path, err)
	}
	r = Media{
		Path:      path,
		Hash:      fmt.Sprintf("%x", hash.Sum(nil)),
		Extension: ext,
		Time:      t,
	}
	return r, nil
}

func Merge(cs []<-chan Media) <-chan Media {
	out := make(chan Media)
	var wg sync.WaitGroup
	output := func(c <-chan Media) {
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

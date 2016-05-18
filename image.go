package arrange

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

type Media struct {
	Path string
}

func (m Media) Move(root string) error {
	return errors.New("NYI")
}

type Image struct {
	Path  string
	Hash  string
	Year  string
	Month string
	Time  string
}

func (im Image) Move(root string) error {
	f, err := os.Open(im.Path)
	if err != nil {
		return fmt.Errorf("problem opening jpg file: %v", err)
	}
	defer f.Close()

	content := filepath.Join(root, "content", im.Hash[:2], im.Hash[2:]+".jpg")

	if _, err := os.Stat(content); !os.IsNotExist(err) {
		return Dup{content}
	}

	out, err := os.Create(content)
	if err != nil {
		return fmt.Errorf("could not create output file: %v", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, f); err != nil {
		return fmt.Errorf("trouble copying file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "date", im.Year, im.Month), 0755); err != nil {
		return fmt.Errorf("problem creating date directory: %v", err)
	}

	date := filepath.Join(root, "date", im.Year, im.Month, im.Time)
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

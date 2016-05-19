package arrange

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Media struct {
	Path      string
	Hash      string
	Extension string
	Time      time.Time
}

func (m Media) Move(root string) error {
	f, err := os.Open(m.Path)
	if err != nil {
		return fmt.Errorf("problem opening file %q: %v", m.Path, err)
	}
	defer f.Close()

	content := filepath.Join(root, "content", m.Hash[:2], m.Hash[2:]+m.Extension)

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

	year := fmt.Sprintf("%04d", m.Time.Year())
	month := fmt.Sprintf("%02d", m.Time.Month())
	ts := fmt.Sprintf("%d", m.Time.UnixNano())

	if err := os.MkdirAll(filepath.Join(root, "date", year, month), 0755); err != nil {
		return fmt.Errorf("problem creating date directory: %v", err)
	}

	date := filepath.Join(root, "date", year, month, ts)
	name := date + m.Extension
	for i := 0; i < 10000; i++ {
		if _, err := os.Stat(name); os.IsNotExist(err) {
			break
		}
		name = fmt.Sprintf("%s_%04d%s", date, i, m.Extension)
	}

	// TODO: or maybe symlinking? (issue #2)
	// rel := filepath.Join("..", "..", "..", "content", j.hash[:2], j.hash[2:]+m.Extension)
	// return os.Symlink(rel, name)
	return os.Link(content, name)
}

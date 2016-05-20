package arrange

import "fmt"

// NotMedia is for unkown filetypes.
type NotMedia struct {
	Path string
}

func (nm NotMedia) Error() string {
	return fmt.Sprintf("not media: %q", nm.Path)
}

// Dup indicates a file with duplicate content.
type Dup struct {
	Path string
}

func (d Dup) Error() string {
	return fmt.Sprintf("dup: %q", d.Path)
}

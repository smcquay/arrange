package arrange

import "fmt"

type NotMedia struct {
	Path string
}

func (nm NotMedia) Error() string {
	return fmt.Sprintf("not media: %q", nm.Path)
}

type Dup struct {
	Path string
}

func (d Dup) Error() string {
	return fmt.Sprintf("dup: %q", d.Path)
}

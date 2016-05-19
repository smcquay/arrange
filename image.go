package arrange

import (
	"fmt"
	"io"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

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

package arrange

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileMove(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp, err := ioutil.TempDir("", "arrange-tests-")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.RemoveAll(tmp)
	}()
	if err := PrepOutput(tmp); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path     string
		expected []string
		ts       time.Time
	}{
		{
			path: filepath.Join(wd, "testdata", "lenna.png"),
			expected: []string{
				"date/2012/10/1350815400000000000.png",
				"content/81/4a0034f5549e957ee61360d87457e5.png",
			},
			ts: time.Date(2012, 10, 21, 10, 30, 0, 0, time.UTC),
		},
		{
			path: filepath.Join(wd, "testdata", "valid.jpg"),
			expected: []string{
				"date/2012/10/1350815400000000001.jpg",
				"content/77/d2a6bf840622331df62963174df72d.jpg",
			},
			ts: time.Date(2012, 10, 21, 10, 30, 0, 1, time.UTC),
		},
		{
			path: filepath.Join(wd, "testdata", "stott.gif"),
			expected: []string{
				"date/2012/10/1350815400000000001.gif",
				"content/9e/f476f6e1ee9ecf64ca443eb63e4655.gif",
			},
			ts: time.Date(2012, 10, 21, 10, 30, 0, 1, time.UTC),
		},
		{
			path: filepath.Join(wd, "testdata", "exif-decode-error.jpg"),
			expected: []string{
				"date/2006/03/1143508767000000000.jpg",
				"content/4e/08875f21249becbc18bc978b35a9b4.jpg",
			},
			ts: time.Date(2012, 10, 21, 10, 30, 0, 2, time.UTC),
		},
		{
			path: filepath.Join(wd, "testdata", "no-exif-but-good.jpg"),
			expected: []string{
				"date/2012/10/1350815400000000003.jpg",
				"content/a2/7735c61b375da9012a33dbdb548274.jpg",
			},
			ts: time.Date(2012, 10, 21, 10, 30, 0, 3, time.UTC),
		},
		{
			path: filepath.Join(wd, "testdata", "bad.mov"),
			expected: []string{
				"date/2012/10/1350815400000000004.mov",
				"content/d4/1d8cd98f00b204e9800998ecf8427e.mov",
			},
			ts: time.Date(2012, 10, 21, 10, 30, 0, 4, time.UTC),
		},
	}
	for _, test := range tests {
		if err := os.Chtimes(test.path, test.ts, test.ts); err != nil {
			t.Fatalf("chtime fail: %v", err)
		}
		m, err := _parse(test.path)
		if err != nil {
			t.Fatalf("problem parsing known good png: %v", err)
		}
		if err := m.Move(tmp); err != nil {
			t.Fatalf("problem moving file into place: %v", err)
		}
		for _, p := range test.expected {
			if _, err := os.Stat(filepath.Join(tmp, p)); os.IsNotExist(err) {
				t.Errorf("could not find expected file %q: %v", p, err)
			}
		}
	}
}

func TestBadFiles(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp, err := ioutil.TempDir("", "arrange-tests-")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.RemoveAll(tmp)
	}()
	if err := PrepOutput(tmp); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		path string
	}{
		{
			path: filepath.Join(wd, "testdata", "bad-link"),
		},
		{
			path: filepath.Join(wd, "testdata", "not.a.jpg"),
		},
		{
			path: filepath.Join(wd, "testdata", "not.a.png"),
		},
		{
			path: filepath.Join(wd, "testdata", "too-many-links.jpg"),
		},
	}
	for _, test := range tests {
		_, err := _parse(test.path)
		if err == nil {
			t.Fatalf("should have had an error in parse of %q, got nil", test.path)
		}
	}
}

func TestMoveCollision(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp, err := ioutil.TempDir("", "arrange-tests-")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.RemoveAll(tmp)
	}()
	if err := PrepOutput(tmp); err != nil {
		t.Fatal(err)
	}
	ts := time.Date(2012, 10, 21, 10, 30, 0, 0, time.UTC)
	media := []Media{
		Media{
			Path:      filepath.Join(wd, "testdata", "a.mov"),
			Hash:      "60b725f10c9c85c70d97880dfe8191b3",
			Time:      ts,
			Extension: ".mov",
		},
		Media{
			Path:      filepath.Join(wd, "testdata", "b.mov"),
			Hash:      "3b5d5c3712955042212316173ccf37be",
			Time:      ts,
			Extension: ".mov",
		},
		Media{
			Path:      filepath.Join(wd, "testdata", "c.mov"),
			Hash:      "2cd6ee2c70b0bde53fbe6cac3c8b8bb1",
			Time:      ts,
			Extension: ".mov",
		},
	}

	for _, m := range media {
		if err := m.Move(tmp); err != nil {
			t.Fatalf("move: %v", err)
		}
	}

	expected := []string{
		"date/2012/10/1350815400000000000_0001.mov",
		"date/2012/10/1350815400000000000_0000.mov",
		"date/2012/10/1350815400000000000.mov",
	}
	for _, p := range expected {
		if _, err := os.Stat(filepath.Join(tmp, p)); os.IsNotExist(err) {
			t.Errorf("could not find expected file %q: %v", p, err)
		}
	}
}

func TestSundry(t *testing.T) {
	fmt.Sprintf("%v", NotMedia{"hi"})
	fmt.Sprintf("%v", Dup{"hi"})
}

func TestFlow(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp, err := ioutil.TempDir("", "arrange-tests-")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.RemoveAll(tmp)
	}()
	if err := PrepOutput(tmp); err != nil {
		t.Fatal(err)
	}

	files := []struct {
		path string
		ts   time.Time
	}{
		{
			path: filepath.Join(wd, "testdata", "a.mov"),
			ts:   time.Date(2012, 10, 21, 10, 30, 0, 0, time.UTC),
		},
		{
			path: filepath.Join(wd, "testdata", "b.mov"),
			ts:   time.Date(2012, 10, 21, 10, 30, 0, 1, time.UTC),
		},
		{
			path: filepath.Join(wd, "testdata", "c.mov"),
			ts:   time.Date(2012, 10, 21, 10, 30, 0, 2, time.UTC),
		},
		{
			path: filepath.Join(wd, "testdata", "bad.mov"),
			ts:   time.Date(2012, 10, 21, 10, 30, 0, 3, time.UTC),
		},

		{
			path: filepath.Join(wd, "testdata", "no-exif-but-good.jpg"),
			ts:   time.Date(2012, 10, 21, 10, 30, 0, 0, time.UTC),
		},
		{
			path: filepath.Join(wd, "testdata", "exif-decode-error.jpg"),
			ts:   time.Date(2012, 10, 21, 10, 30, 0, 1, time.UTC),
		},
		{
			path: filepath.Join(wd, "testdata", "valid.jpg"),
			ts:   time.Date(2012, 10, 21, 10, 30, 0, 2, time.UTC),
		},

		{
			path: filepath.Join(wd, "testdata", "stott.gif"),
			ts:   time.Date(2012, 10, 21, 10, 30, 0, 0, time.UTC),
		},
		{
			path: filepath.Join(wd, "testdata", "lenna.png"),
			ts:   time.Date(2012, 10, 21, 10, 30, 0, 0, time.UTC),
		},
	}
	for _, f := range files {
		if err := os.Chtimes(f.path, f.ts, f.ts); err != nil {
			t.Errorf("failure to chtime: %v", err)
		}
	}

	work := Source(filepath.Join(wd, "testdata"))
	streams := []<-chan Media{}

	for w := 0; w < 4; w++ {
		streams = append(streams, Parse(work))
	}

	for err := range Move(Merge(streams), tmp) {
		if err != nil {
			t.Errorf("unexpected error: %v")
		}
	}

	expected := []string{
		"date/2012/10/1350815400000000000.mov",
		"date/2012/10/1350815400000000001.mov",
		"date/2012/10/1350815400000000002.mov",
		"date/2012/10/1350815400000000003.mov",

		"date/2012/10/1350815400000000000.jpg",
		"date/2012/10/1350815400000000002.jpg",
		// this one has valid exif
		"date/2006/03/1143508767000000000.jpg",

		"date/2012/10/1350815400000000000.gif",
		"date/2012/10/1350815400000000000.png",
	}

	for _, p := range expected {
		if _, err := os.Stat(filepath.Join(tmp, p)); os.IsNotExist(err) {
			t.Errorf("could not find expected file %q: %v", p, err)
		}
	}
}

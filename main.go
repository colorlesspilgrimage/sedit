package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"golang.org/x/term"
)

type lineBuf map[int64][]byte

const FrontBuf = true
const BackBuf = false

type DisplayBuf struct {
	FrontBuf [][]byte
	BackBuf  [][]byte
	Height   int
	Width    int
}

func NewDisplayBuf(tfd int, f *os.File) *DisplayBuf {
	w, h, e := term.GetSize(tfd)
	s := bufio.NewScanner(f)
	l := 0
	ll := 0
	var b [][]byte

	checkErr(e)

	for s.Scan() {
		l++
	}

	checkErr(s.Err())

	_, e = f.Seek(0, io.SeekStart)
	s = bufio.NewScanner(f)

	for s.Scan() {
		ln := s.Text()
		if len(ln) > ll {
			ll = len(ln)
		}
	}

	checkErr(s.Err())

	_, e = f.Seek(0, io.SeekStart)
	checkErr(e)

	b = make([][]byte, l)
	for i := range b {
		b[i] = make([]byte, ll)
	}

	return &DisplayBuf{
		FrontBuf: b,
		BackBuf:  b,
		Height:   h,
		Width:    w,
	}
}

// when we load the buffer, we actually want to load the total amount of data from the file read in.
// currently this only loads what can be seen on the display when the program is run.
// we only care about the width and height when we talk about drawing the buffer
// this will also require some changes to our creation function, as it will also need to know the size of the file its working with
// in order to inform what we need to size the buffer to. there is some question in my mind as to how big the buffers need to be though -
// there is no need for the w & h slices of the buffer to each be big enough to store the entire file.
func (db *DisplayBuf) LoadBuf(b *[]byte, w bool) {
	switch w {
	case true:
		// these functions cause a panic due to OOB access, probably because we are not checking that the height is also less than the maximum lines we
		// allocate against.
		for i := 0; i < db.Height; i++ {
			for j := 0; (j < db.Width) && (j < len(*b)); j++ {
				db.FrontBuf[i][j] = (*b)[j]
			}
		}
	case false:
		for i := 0; i < db.Height; i++ {
			for j := 0; (j < db.Width) && (j < len(*b)); j++ {
				db.BackBuf[i][j] = (*b)[j]
			}
		}
	}

}

func (db *DisplayBuf) InitBuf(f *os.File) {
	b := make([]byte, func() int {
		s, e := f.Stat()
		checkErr(e)
		return int(s.Size())
	}())

	_, e := f.Read(b)
	checkErr(e)

	db.LoadBuf(&b, FrontBuf)
	db.LoadBuf(&b, BackBuf)
}

func (db *DisplayBuf) DrawBuf(w bool) {
	switch w {
	case true:
		for i := 0; i < db.Height; i++ {
			for j := 0; j < db.Width; j++ {
				fmt.Printf("%c", db.FrontBuf[i][j])
			}
		}
	case false:
		for i := 0; i < db.Height; i++ {
			for j := 0; j < db.Width; j++ {
				fmt.Printf("%c", db.FrontBuf[i][j])
			}
		}
	}
}

func checkErr(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

/*
gap buffers - all text for the current piece is loaded into a buffer, and a gap is inserted at the point in the buffer where the editing cursor should start.
as the cursor is moved, data is moved in the buffer relative to the direction of the cursors movement, essentialy causing the gap to "follow" the
cursor. the gap always expands itself when new data is added to maintain a gap of a certain size for continual editing, and it is reset when the data
is written to the file on a save. it expands when text is deleted. the concept of "lines" is not visible to the gap buffer, meaning that we only need one actual buffer
that needs to be loaded with each line from the piece table.

piece table - when we read the file, we break it into points to each "piece", in this case, each line as defined by reading a chunk of data until a '\n' character
is found. once one is found, we create the next piece of data, and repeat until EOF.  we can then assign line numbers to each of these pieces, and use it to draw
them on the screen. as we move from line to line, the gap buffer is loaded with a copy of the data from the current line in the piece table, and we redraw the
screen with the copied data. the screen will then reflect the edited data from the buffer. on write, we swap back to displaying the piece table data after we
re-read the file.

double buffer display - two buffers to allow for smooth redrawing. [width][height]display or somthing like that. each time the user enters input we update
the back buffer and then copy differences to the front buffer to smoothly update the display.

moving the cursor - when the user enters input, we also need to move the cursor and redraw the double buffer to reflect changes.
long term efficiency gains here could be from only redrawing the buffer when the cursor moves enough to cause a new line to need to be drawn.

*/

func main() {

	var e error
	var f *os.File
	a := os.Args[1:]
	in := bufio.NewScanner(os.Stdin)
	tfd := int(os.Stdout.Fd())
	var db *DisplayBuf

	// make sure we are running in a terminal
	if !term.IsTerminal(tfd) {
		fmt.Println("not running in a terminal!")
		return
	}

	if len(a) <= 0 {
		// eventually we will want this to create an unnamed file and still open a buffer
		// it will need to create a temporary hidden waste file to store the buffer data in until
		// it is written to disk with a real file name or discarded.
		// the temp buffer will need to be dynamically expended with append() to ensure we have space for
		// the data we are writing as there wont be any "pre-defined" line sizes fom a read file.
		return
	}

	f, e = os.OpenFile(a[0], os.O_RDWR|os.O_CREATE, 0644)
	defer f.Close()
	checkErr(e)

	db = NewDisplayBuf(tfd, f)

	db.InitBuf(f)
	db.DrawBuf(FrontBuf)

	for {
		in.Scan()
		break
	}
}

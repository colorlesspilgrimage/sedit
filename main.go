package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// custom type to make it easier to read when we are referencing specifically a piece of the piece table
type piece []byte
type lineBuf map[int64]piece

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
	var c []byte
	var l int64
	var i int64
	a := os.Args[1:]
	var s *bufio.Scanner
	var pt lineBuf = make(map[int64]piece)
	in := bufio.NewScanner(os.Stdin)

	if len(a) <= 0 {
		// eventually we will want this to create an unnamed file and still open a buffer
		// it will need to create a temporary hidden waste file to store the buffer data in until
		// it is written to disk with a real file name or discarded.
		// the temp buffer will need to be dynamically expended with append() to ensure we have space for
		// the data we are writing as there wont be any "pre-defined" line sizes fom a read file.
		return
	}

	f, e = os.OpenFile(a[0], os.O_RDWR|os.O_CREATE, 0644)
	checkErr(e)

	l = (func() int64 {
		s, e := f.Stat()
		checkErr(e)
		return s.Size()
	})()

	c = make([]byte, l)
	_, e = f.Read(c)
	checkErr(e)

	s = bufio.NewScanner(strings.NewReader(string(c)))
	for s.Scan() {
		pt[i] = []byte(s.Text())
		i++
	}
	checkErr(s.Err())

	fmt.Printf("sediting file: %s, enter command:\n", a[0])

	for {
		in.Scan()
		break
	}

	e = f.Close()
	checkErr(e)
}

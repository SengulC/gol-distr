package gol

import (
	"strconv"
	"uk.ac.bris.cs/gameoflife/util"
)

type distributorChannels struct {
	events     chan<- Event
	ioCommand  chan<- ioCommand
	ioIdle     <-chan bool
	ioFilename chan<- string
	ioOutput   chan<- uint8
	ioInput    <-chan uint8
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	// TODO: Create a 2D slice to store the world.
	// figure out the name of the file from the params, send it down channel
	// after sending down appropriate command, start io goroutine
	name := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight)
	// name := fmt.Sprintf("%dx%d", p.ImageWidth, p.ImageHeight)
	c.ioCommand <- ioInput
	c.ioFilename <- name

	worldIn := make([][]byte, p.ImageHeight)
	for i := range worldIn {
		worldIn[i] = make([]byte, p.ImageWidth)
	}

	// get image byte by byte and store in: worldIn
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			worldIn[row][col] = <-c.ioInput
			// fmt.Println(worldIn[row][col])
		}
	}

	/*
		SEND DATA OVER TO GOL WORKER
		...
		ONCE RECEIVED CONTINUE
	*/

	// count final worldOut's state
	var count int
	var cells []util.Cell
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			if worldOut[row][col] == 255 {
				c := util.Cell{X: col, Y: row}
				cells = append(cells, c)
				count++
			}
		}
	}

	// TODO: Report the final state using FinalTurnCompleteEvent.
	// pass it down events channel (list of alive cells)
	c.events <- FinalTurnComplete{p.Turns, cells}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle

	c.events <- StateChange{p.Turns, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

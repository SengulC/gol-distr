package gol

import (
	"flag"
	"fmt"
	"net/rpc"
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

var UpdateHandler = "UpdateOperations.UpdateBoard"

type Response struct {
	World          [][]byte
	Cells          []util.Cell
	AliveCellCount int
}

type Request struct {
	World [][]byte
	P     Params
}

var response = new(Response)

func makeCall(client *rpc.Client, world [][]byte, p Params) {
	request := Request{World: world, P: p}
	client.Call(UpdateHandler, request, response)
	fmt.Println("Responded")
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	// TODO: Create a 2D slice to store the world.
	name := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight)
	c.ioCommand <- ioInput
	c.ioFilename <- name

	// initialise worldIn
	worldIn := make([][]byte, p.ImageHeight)
	for i := range worldIn {
		worldIn[i] = make([]byte, p.ImageWidth)
	}

	// get image byte by byte and store in: worldIn
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			worldIn[row][col] = <-c.ioInput
		}
	}

	// TODO: Execute all turns of the Game of Life.

	server := flag.String("server", "127.0.0.1:8050", "IP:port string to connect to as server")
	flag.Parse()
	client, _ := rpc.Dial("tcp", *server)
	// connected to server
	defer client.Close()

	makeCall(client, worldIn, p)
	// pass worldIn to server

	// TODO: Report the final state using FinalTurnCompleteEvent.
	// get back info from server

	c.events <- FinalTurnComplete{p.Turns, response.Cells}
	// where cells: alive cells!

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{p.Turns, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

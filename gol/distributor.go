package gol

import (
	"flag"
	"fmt"
	"net/rpc"
	"strconv"
	"time"
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

var UpdateHandler = "UpdateOperations.Update"

type Response struct {
	World          [][]byte
	Cells          []util.Cell
	Turns          int
	AliveCellCount int
}

type Request struct {
	World [][]byte
	P     Params
}

var server = flag.String("server", "127.0.0.1:8050", "IP:port string to connect to as server")

//var flagBool = false

func makeWorld(world [][]byte) [][]byte {
	world2 := make([][]byte, len(world))
	for col := 0; col < len(world); col++ {
		world2[col] = make([]byte, len(world))
	}
	return world2
}

func makeCall(client *rpc.Client, world [][]byte, p Params, c distributorChannels, response *Response) {
	timeOver := time.NewTicker(2 * time.Second)
	//fmt.Println("entered makeCall")
	request := Request{World: world, P: p}

	select {
	case <-timeOver.C:
		fmt.Println("2 secs have past!")
		c.events <- AliveCellsCount{response.Turns, response.AliveCellCount}
	default:
		fmt.Println("DEFT: making call to update handler")
		client.Call(UpdateHandler, request, response)
		//client.Go(UpdateHandler, request, response, make(chan *Call, 1))
	}
	//fmt.Println("Responded")
}

// distributor divides the work between workers and interacts with other goroutines.
func distributor(p Params, c distributorChannels) {
	// TODO: Create a 2D slice to store the world.
	//fmt.Println("distributor")
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

	//fmt.Println("trying to connect to server")
	//if flagBool == false {
	//	server = flag.String("server", "127.0.0.1:8050", "IP:port string to connect to as server")
	//	flag.Parse()
	//	flagBool = true
	//}
	client, _ := rpc.Dial("tcp", *server)
	//fmt.Println("connected to server")
	defer client.Close()

	var response = new(Response)
	response.World = makeWorld(response.World)
	makeCall(client, worldIn, p, c, response)
	// pass worldIn to server

	// TODO: Report the final state using FinalTurnCompleteEvent.
	// get back info from server

	c.ioCommand <- ioOutput
	c.ioFilename <- name + "x" + strconv.Itoa(p.Turns)
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			c.ioOutput <- response.World[row][col]
		}
	}

	c.events <- FinalTurnComplete{p.Turns, response.Cells}
	// where cells: alive cells!

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{p.Turns, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

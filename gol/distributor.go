package gol

import (
	"flag"
	"fmt"
	"net/rpc"
	"os"
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
	keyPresses <-chan rune
}

var UpdateHandler = "UpdateOperations.Update"
var TickerHandler = "UpdateOperations.Ticker"
var SaveHandler = "UpdateOperations.Save"
var PauseHandler = "UpdateOperations.Pause"
var ContinueHandler = "UpdateOperations.Continue"
var QuitHandler = "UpdateOperations.Quit"
var KillHandler = "UpdateOperations.Kill"
var PreservedHandler = "UpdateOperations.FetchPreserved"

type Response struct {
	World          [][]byte
	AliveCells     []util.Cell
	CompletedTurns int
	AliveCellCount int
	Preserved      bool
}

type Request struct {
	World  [][]byte
	P      Params
	Events chan<- Event
}

// to be changed to AWS node when AWS instance is instantiated
var server = flag.String("server", "127.0.0.1:8050", "IP:port string to connect to as server")

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
			//if worldIn[row][col] == 255 {
			//	c.events <- CellFlipped{0, util.Cell{X: col, Y: row}}
			//}
		}
	}
	// TODO: Execute all turns of the Game of Life.

	client, _ := rpc.Dial("tcp", *server)
	defer client.Close()

	var response = new(Response)
	var tickerRes = new(Response)
	var saveRes = new(Response)
	var pauseRes = new(Response)
	var preservedRes = new(Response)
	var goCall *rpc.Call
	request := Request{World: worldIn, P: p}

	client.Call(PreservedHandler, Request{}, preservedRes)
	if preservedRes.Preserved == true {
		goCall = client.Go(ContinueHandler, Request{}, pauseRes, nil)
		<-goCall.Done
	} else {
		goCall = client.Go(UpdateHandler, request, response, nil)
	}

	timeOver := time.NewTicker(2 * time.Second)
	var key rune
	paused := false
	quit := false
L:
	for {
		select {
		case key = <-c.keyPresses:
			switch key {
			case 'p':
				if !paused {
					paused = true
					client.Call(PauseHandler, Request{}, pauseRes)
					c.events <- StateChange{pauseRes.CompletedTurns, Paused}
					fmt.Println("Paused. Current turn:", pauseRes.CompletedTurns)
				} else {
					paused = false
					client.Call(ContinueHandler, Request{}, pauseRes)
					c.events <- StateChange{pauseRes.CompletedTurns, Executing}
					fmt.Println("Continuing!")
				}
			case 's':
				c.ioCommand <- ioOutput
				c.ioFilename <- name + "x" + strconv.Itoa(p.Turns)
				fmt.Println("Saving...", name+"x"+strconv.Itoa(p.Turns)+"")
				client.Call(SaveHandler, request, saveRes)
				for row := 0; row < p.ImageHeight; row++ {
					for col := 0; col < p.ImageWidth; col++ {
						c.ioOutput <- saveRes.World[row][col]
					}
				}
			case 'k':
				quit = true
				c.events <- StateChange{p.Turns, Quitting}
				client.Call(KillHandler, Request{}, Response{})
				fmt.Println("KILLING")
				os.Exit(0)
			case 'q':
				c.events <- StateChange{p.Turns, Quitting}
				client.Call(QuitHandler, Request{}, Response{})
				os.Exit(0)
			}
		case <-goCall.Done:
			break L
		case <-timeOver.C:
			if !paused {
				client.Call(TickerHandler, Request{}, tickerRes)
				c.events <- AliveCellsCount{CompletedTurns: tickerRes.CompletedTurns, CellsCount: tickerRes.AliveCellCount}
			}
		}
	}

	// TODO: Report the final state using FinalTurnCompleteEvent.

	if !quit {
		c.ioCommand <- ioOutput
		c.ioFilename <- name + "x" + strconv.Itoa(p.Turns)
		for row := 0; row < p.ImageHeight; row++ {
			for col := 0; col < p.ImageWidth; col++ {
				c.ioOutput <- response.World[row][col]
			}
		}
	}

	c.events <- FinalTurnComplete{p.Turns, response.AliveCells}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{p.Turns, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

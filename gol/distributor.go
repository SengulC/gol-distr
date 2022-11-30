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
	keyPresses <-chan rune
}

var BrokerGOLHandler = "BrokerOperations.BrokerGOL"
var BrokerTickerHandler = "BrokerOperations.BrokerTicker"
var BrokerPauseHandler = "BrokerOperations.BrokerPause"
var BrokerSaveHandler = "BrokerOperations.BrokerSave"
var BrokerContinueHandler = "BrokerOperations.BrokerContinue"

type Response struct {
	World          [][]byte
	AliveCells     []util.Cell
	CompletedTurns int
	AliveCellCount int
}

type Request struct {
	World  [][]byte
	P      Params
	Events chan<- Event
}

//var brokerServer = flag.String("brokerServer", "3.91.54.94:8050", "IP:port string to connect to as brokerServer")

var brokerServer = flag.String("brokerServer", "127.0.0.1:8050", "IP:port string to connect to as brokerServer")

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
			//if worldIn[row][col] == 255 {
			//	c.events <- CellFlipped{0, util.Cell{X: col, Y: row}}
			//}
		}
	}

	// TODO: Execute all turns of the Game of Life.

	client, _ := rpc.Dial("tcp", *brokerServer)
	defer client.Close()

	request := Request{World: worldIn, P: p}
	var response = new(Response)
	var tickerRes = new(Response)
	var pauseRes = new(Response)
	var saveRes = new(Response)

	goCall := client.Go(BrokerGOLHandler, request, response, nil)

	var key rune
	paused := false
	timeOver := time.NewTicker(2 * time.Second)

L:
	for {
		select {
		case key = <-c.keyPresses:
			switch key {
			case 'p':
				if !paused {
					paused = true
					client.Call(BrokerPauseHandler, Request{}, pauseRes)
					fmt.Println("PAUSED. Current turn:", pauseRes.CompletedTurns)
				} else {
					paused = false
					fmt.Println("Continuing...")
					client.Call(BrokerContinueHandler, Request{}, pauseRes)
				}
			case 's':
				fmt.Println("Saving...")
				c.ioCommand <- ioOutput
				c.ioFilename <- name + "x" + strconv.Itoa(p.Turns)
				client.Call(BrokerSaveHandler, request, saveRes)
				for row := 0; row < p.ImageHeight; row++ {
					for col := 0; col < p.ImageWidth; col++ {
						c.ioOutput <- saveRes.World[row][col]
					}
				}
			case 'q':
			case 'k':
			}
		case <-goCall.Done:
			break L
		case <-timeOver.C:
			if !paused {
				client.Call(BrokerTickerHandler, Request{}, tickerRes)
				c.events <- AliveCellsCount{CompletedTurns: tickerRes.CompletedTurns, CellsCount: tickerRes.AliveCellCount}
			}
		}
	}

	// TODO: Report the final state using FinalTurnCompleteEvent.
	// get back info from brokerServer

	fmt.Println(response.World)
	fmt.Println("starting output")
	c.ioCommand <- ioOutput
	c.ioFilename <- name + "x" + strconv.Itoa(p.Turns)
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			c.ioOutput <- response.World[row][col]
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

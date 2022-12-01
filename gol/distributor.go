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

var server = flag.String("server", "127.0.0.1:8050", "IP:port string to connect to as server")

//var server = flag.String("server", "127.0.0.1:8050", "IP:port string to connect to as server")

//var flagBool = false

func makeMatrix(world [][]byte) [][]byte {
	world2 := make([][]byte, len(world))
	for col := 0; col < len(world); col++ {
		world2[col] = make([]byte, len(world))
	}
	return world2
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
			//if worldIn[row][col] == 255 {
			//	c.events <- CellFlipped{0, util.Cell{X: col, Y: row}}
			//}
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
	defer client.Close()

	var response = new(Response)
	var tickerRes = new(Response)
	var saveRes = new(Response)
	var pauseRes = new(Response)
	request := Request{World: worldIn, P: p}
	//response.World = makeWorld(response.World)

	//var key rune
	//timeOver := time.NewTicker(2 * time.Second)
	//select {
	//case key = <-c.keyPresses:
	//	switch key {
	//	case 's':
	//		//save
	//		//fmt.Println("Saving")
	//		//c.ioCommand <- ioOutput
	//		//name := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight)
	//		//c.ioFilename <- name + "x" + strconv.Itoa(p.Turns)
	//		//for row := 0; row < p.ImageHeight; row++ {
	//		//	for col := 0; col < p.ImageWidth; col++ {
	//		//		c.ioOutput <- response.World[row][col]
	//		//	}
	//		//}
	//		//save()
	//		//case 'q':
	//		//	//Close the controller client program without causing an error on the GoL server.
	//		//	//A new controller should be able to take over interaction with the GoL engine.
	//		//	//Note that you are free to define the nature of how a new controller can take over interaction.
	//		//	//Most likely the state will be reset.
	//		//case 'k':
	//		//	//All components of the distributed system are shut down cleanly,
	//		//	//& the system outputs a PGM image of the latest state
	//		//case 'p':
	//		//	//Pause the processing on the AWS node and have the controller print the current turn
	//		//	//If p is pressed again resume the processing and have the controller print "Continuing".
	//	}
	//case <-timeOver.C:
	//	ticker(client, response, request)
	//default:
	goCall := client.Go(UpdateHandler, request, response, nil)

	//fmt.Println(response.AliveCells)
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
					fmt.Println("Paused. Current turn:", pauseRes.CompletedTurns)
				} else {
					paused = false
					client.Call(ContinueHandler, Request{}, pauseRes)
				}
			case 's':
				fmt.Println("Saving...")
				c.ioCommand <- ioOutput
				c.ioFilename <- name + "x" + strconv.Itoa(p.Turns)
				client.Call(SaveHandler, request, saveRes)
				fmt.Println("ON CLIENT", len(saveRes.World))
				for row := 0; row < p.ImageHeight; row++ {
					for col := 0; col < p.ImageWidth; col++ {
						c.ioOutput <- saveRes.World[row][col]
					}
				}
				fmt.Println("Saving...", saveRes.CompletedTurns)
			case 'k':
				quit = true
				client.Call(KillHandler, Request{}, Response{})
				break
			case 'q':
				client.Call(QuitHandler, Request{}, Response{})
				os.Exit(0)
				break
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
	// get back info from server

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

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

var UpdateHandler = "UpdateOperations.Update"
var TickerHandler = "UpdateOperations.Ticker"

type Response struct {
	World          [][]byte
	AliveCells     []util.Cell
	CompletedTurns int
	AliveCellCount int
}

type Request struct {
	World    [][]byte
	P        Params
	Events   chan<- Event
	TurnZero chan bool
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

func makeCall(client *rpc.Client, p Params, events chan<- Event, response *Response, request Request) {
	//timeOver := time.NewTicker(2 * time.Second)
	//for {
	//	select {
	//	//case key = <-c.keyPresses:
	//	//	switch key {
	//	//	case 's':
	//	//		//save
	//	//		fmt.Println("Saving")
	//	//		c.ioCommand <- ioOutput
	//	//		name := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight)
	//	//		c.ioFilename <- name + "x" + strconv.Itoa(p.Turns)
	//	//		for row := 0; row < p.ImageHeight; row++ {
	//	//			for col := 0; col < p.ImageWidth; col++ {
	//	//				c.ioOutput <- response.World[row][col]
	//	//			}
	//	//		}
	//	//	case 'q':
	//	//		//Close the controller client program without causing an error on the GoL server.
	//	//		//A new controller should be able to take over interaction with the GoL engine.
	//	//		//Note that you are free to define the nature of how a new controller can take over interaction.
	//	//		//Most likely the state will be reset.
	//	//	case 'k':
	//	//		//All components of the distributed system are shut down cleanly,
	//	//		//& the system outputs a PGM image of the latest state
	//	//	case 'p':
	//	//		//Pause the processing on the AWS node and have the controller print the current turn
	//	//		//If p is pressed again resume the processing and have the controller print "Continuing".
	//	//	}
	//	case <-timeOver.C:
	//		fmt.Println("2 seconds have passed! making ticker call")
	//		goCall = client.Go(TickerHandler, request, response, nil)
	//		<-goCall.Done
	//	default:
	//		fmt.Println("DEFAULT: making call to update handler")
	//		goCall = client.Go(UpdateHandler, request, response, nil)
	//		<-goCall.Done
	//		c.events <- FinalTurnComplete{p.Turns, response.AliveCells}
	//	}
	//}
	//fmt.Println("Responded")
	fmt.Println("DEFAULT: making call to update handler")
	goCall := client.Go(UpdateHandler, request, response, nil)
	<-goCall.Done
	events <- FinalTurnComplete{p.Turns, response.AliveCells}
}

func ticker(client *rpc.Client, response *Response, request Request) {
	//timeOver := time.NewTicker(2 * time.Second)
	//for {
	//	select {
	//	//case key = <-c.keyPresses:
	//	//	switch key {
	//	//	case 's':
	//	//		//save
	//	//		fmt.Println("Saving")
	//	//		c.ioCommand <- ioOutput
	//	//		name := strconv.Itoa(p.ImageWidth) + "x" + strconv.Itoa(p.ImageHeight)
	//	//		c.ioFilename <- name + "x" + strconv.Itoa(p.Turns)
	//	//		for row := 0; row < p.ImageHeight; row++ {
	//	//			for col := 0; col < p.ImageWidth; col++ {
	//	//				c.ioOutput <- response.World[row][col]
	//	//			}
	//	//		}
	//	//	case 'q':
	//	//		//Close the controller client program without causing an error on the GoL server.
	//	//		//A new controller should be able to take over interaction with the GoL engine.
	//	//		//Note that you are free to define the nature of how a new controller can take over interaction.
	//	//		//Most likely the state will be reset.
	//	//	case 'k':
	//	//		//All components of the distributed system are shut down cleanly,
	//	//		//& the system outputs a PGM image of the latest state
	//	//	case 'p':
	//	//		//Pause the processing on the AWS node and have the controller print the current turn
	//	//		//If p is pressed again resume the processing and have the controller print "Continuing".
	//	//	}
	//	case <-timeOver.C:
	//		fmt.Println("2 seconds have passed! making ticker call")
	//		goCall = client.Go(TickerHandler, request, response, nil)
	//		<-goCall.Done
	//	default:
	//		fmt.Println("DEFAULT: making call to update handler")
	//		goCall = client.Go(UpdateHandler, request, response, nil)
	//		<-goCall.Done
	//		c.events <- FinalTurnComplete{p.Turns, response.AliveCells}
	//	}
	//}
	//fmt.Println("Responded")
	timeOver := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-timeOver.C:
			fmt.Println("2 SECONDS HAVE PASSED: making call to ticker handler")
			fmt.Println("res.compTurns=", response.CompletedTurns)

			goCall := client.Go(TickerHandler, request, response, nil)
			<-goCall.Done
			fmt.Println("RECEIVED from handler!")
		}
	}
}

//func ticker(events chan<- Event, completedTurns, aliveCellCount int, turnZero chan bool) {
//	timeOver := time.Tick(2 * time.Second)
//	for {
//		//<-turnZero
//		select {
//		case <-timeOver:
//			fmt.Println("2 secs have passed!")
//			//fmt.Println("2 secs have passed! AND it's not turn 0!")
//			events <- AliveCellsCount{completedTurns, aliveCellCount}
//		}
//	}
//}

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
	//fmt.Println("connected to server")
	defer client.Close()

	turnZero := make(chan bool, 1)

	var response = new(Response)
	request := Request{World: worldIn, P: p, Events: c.events, TurnZero: turnZero}
	response.World = makeWorld(response.World)

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
	timeOver := time.NewTicker(2 * time.Second)

L:
	for {
		select {
		case <-goCall.Done:
			break L
		case <-timeOver.C:
			client.Call(TickerHandler, request, response)
		}
	}

	//}

	// TODO: Report the final state using FinalTurnCompleteEvent.
	// get back info from server

	c.ioCommand <- ioOutput
	c.ioFilename <- name + "x" + strconv.Itoa(p.Turns)
	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			c.ioOutput <- response.World[row][col]
		}
	}

	//c.events <- FinalTurnComplete{p.Turns, response.AliveCells}

	// Make sure that the Io has finished any output before exiting.
	c.ioCommand <- ioCheckIdle
	<-c.ioIdle
	c.events <- StateChange{p.Turns, Quitting}

	// Close the channel to stop the SDL goroutine gracefully. Removing may cause deadlock.
	close(c.events)
}

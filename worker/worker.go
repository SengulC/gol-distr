package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"sync"
	"time"
	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/util"
)

// server

// UpdateBoard TODO: Update a single iteration
func UpdateBoard(worldIn [][]byte, p gol.Params, events chan<- gol.Event, currentTurn int) [][]byte {
	// worldOut = worldIn
	worldOut := make([][]byte, p.ImageHeight)
	for row := 0; row < p.ImageHeight; row++ {
		worldOut[row] = make([]byte, p.ImageWidth)
		for col := 0; col < p.ImageWidth; col++ {
			worldOut[row][col] = 0
		}
	}

	for row := 0; row < p.ImageHeight; row++ {
		for col := 0; col < p.ImageWidth; col++ {
			// CURRENT ELEMENT AND ITS NEIGHBOR COUNT RESET
			element := worldIn[row][col]
			counter := 0

			// iterate through all neighbors of given element
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					// creates 3x3 matrix w element as centerpiece, but centerpiece is included as well ofc.
					nRow := (row + dx + p.ImageHeight) % p.ImageHeight
					nCol := (col + dy + p.ImageWidth) % p.ImageWidth
					// increment counter if given neighbor is alive
					if worldIn[nRow][nCol] == 255 {
						counter++
						// fmt.Println(counter)
					}
				}
			}

			// if element is alive exclude it from the 3x3 matrix counter
			if element == 255 {
				counter--
			}

			// if element dead, 0
			if element == 0 {
				if counter == 3 {
					worldOut[row][col] = 255
					//events <- gol.CellFlipped{CompletedTurns: currentTurn, Cell: util.Cell{X: col, Y: row}}
				} else {
					worldOut[row][col] = 0
				}
			} else {
				// if element alive, 255
				if counter < 2 {
					worldOut[row][col] = 0
					//events <- gol.CellFlipped{CompletedTurns: currentTurn, Cell: util.Cell{X: col, Y: row}}
				} else if counter > 3 {
					worldOut[row][col] = 0
					//events <- gol.CellFlipped{CompletedTurns: currentTurn, Cell: util.Cell{X: col, Y: row}}
				} else {
					worldOut[row][col] = 255
				}
			}
		}
	}

	return worldOut
}

func calcAliveCellCount(height, width int, world [][]byte) int {
	var count int
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			if world[row][col] == 255 {
				count++
			}
		}
	}
	return count
}

func calcAliveCells(height, width int, world [][]byte) []util.Cell {
	var cells []util.Cell
	for row := 0; row < height; row++ {
		for col := 0; col < width; col++ {
			if world[row][col] == 255 {
				c := util.Cell{X: col, Y: row}
				cells = append(cells, c)
			}
		}
	}
	return cells
}

func makeMatrixOfSameSize(world [][]byte) [][]byte {
	world2 := make([][]byte, len(world))
	for col := 0; col < len(world); col++ {
		world2[col] = make([]byte, len(world))
	}
	return world2
}

func copyMatrix(height, width int, world [][]byte) [][]byte {
	var world2 [][]byte
	for col := 0; col < height; col++ {
		for row := 0; row < width; row++ {
			world2[row][col] = world[row][col]
		}
	}
	return world2
}

func pauseLoop(kP <-chan rune, pause chan bool) {
	fmt.Println("in pauseLoop func")
	for {
		fmt.Println("in the for loop")
		k := <-kP
		fmt.Println("in the for loop2")
		if k == 'p' {
			fmt.Println("another P!")
			pause <- true
			break
		}
		fmt.Println("nothings here yet")
	}
}

type UpdateOperations struct {
	completedTurns int
	aliveCells     int
	mutex          sync.Mutex
	currentWorld   [][]byte
}

func (s *UpdateOperations) Example(req gol.Request, res *gol.Response) (err error) {
	return
}

func (s *UpdateOperations) Ticker(req gol.Request, res *gol.Response) (err error) {
	//fmt.Println("in the ticker method!")
	s.mutex.Lock()

	res.CompletedTurns = s.completedTurns
	//fmt.Println("ticker alive cells:", s.aliveCells)
	res.AliveCellCount = s.aliveCells

	s.mutex.Unlock()
	return
}

func (s *UpdateOperations) Pause(req gol.Request, res *gol.Response) (err error) {
	s.mutex.Lock()
	go pauseLoop(req.KeyPresses, req.Pause)
	fmt.Println("locked and launched go pause loop")
	_ = <-req.Pause
	fmt.Println("got smth from pause chan")
	s.mutex.Unlock()
	res.CompletedTurns = s.completedTurns
	return
}

func (s *UpdateOperations) Save(req gol.Request, res *gol.Response) (err error) {
	fmt.Println("IN SAVE METHOD")
	s.mutex.Lock()
	res.World = makeMatrixOfSameSize(s.currentWorld)
	fmt.Println("ON SERVER", len(res.World))
	fmt.Println("ON SERVER", len(s.currentWorld))
	for col := 0; col < req.P.ImageHeight; col++ {
		for row := 0; row < req.P.ImageWidth; row++ {
			res.World[col][row] = s.currentWorld[col][row]
		}
	}
	s.mutex.Unlock()
	return
}

func (s *UpdateOperations) Update(req gol.Request, res *gol.Response) (err error) {
	fmt.Println("in the upd method")
	if len(req.World) == 0 {
		err = errors.New("world is empty")
		return
	}

	s.currentWorld = make([][]byte, req.P.ImageHeight)
	for row := 0; row < req.P.ImageHeight; row++ {
		s.currentWorld[row] = make([]byte, req.P.ImageWidth)
		for col := 0; col < req.P.ImageWidth; col++ {
			s.currentWorld[row][col] = req.World[row][col]
		}
	}

	turn := 0
	s.completedTurns = 0
	for turn < req.P.Turns {
		s.mutex.Lock()
		s.currentWorld = UpdateBoard(s.currentWorld, req.P, req.Events, turn)
		s.mutex.Unlock()
		s.completedTurns = turn
		//fmt.Println(s.completedTurns)
		s.aliveCells = calcAliveCellCount(req.P.ImageHeight, req.P.ImageWidth, s.currentWorld)
		turn++
	}

	fmt.Println(res.AliveCells)
	s.aliveCells = calcAliveCellCount(req.P.ImageHeight, req.P.ImageWidth, s.currentWorld)
	res.CompletedTurns = turn
	res.World = s.currentWorld
	res.AliveCellCount = s.aliveCells
	res.AliveCells = calcAliveCells(req.P.ImageHeight, req.P.ImageWidth, s.currentWorld)
	return
}

func main() {
	pAddr := flag.String("port", "8050", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&UpdateOperations{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	rpc.Accept(listener)
	// do we need 2 change any of this?
}

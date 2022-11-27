package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
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

type UpdateOperations struct{}

func (s *UpdateOperations) Ticker(req gol.Request, res *gol.Response) (err error) {
	fmt.Println("in the ticker method!")
	if res.CompletedTurns == 0 {
		fmt.Println("umm sorry it's turn", res.CompletedTurns)
		return
	} else {
		req.Events <- gol.AliveCellsCount{CompletedTurns: res.CompletedTurns, CellsCount: calcAliveCellCount(req.P.ImageHeight, req.P.ImageWidth, res.World)}
		return
	}
}

//func (s *UpdateOperations) SaveImage(req gol.Request, res *gol.Response) (err error) {
//	fmt.Println("Saving")
//	req.C.ioCommand <- gol.ioOutput
//	name := strconv.Itoa(req.P.ImageWidth) + "x" + strconv.Itoa(req.P.ImageHeight)
//	req.C.ioFilename <- name + "x" + strconv.Itoa(req.P.Turns)
//	for row := 0; row < req.P.ImageHeight; row++ {
//		for col := 0; col < req.P.ImageWidth; col++ {
//			req.C.ioOutput <- res.World[row][col]
//		}
//	}
//	return
//}

func (s *UpdateOperations) Update(req gol.Request, res *gol.Response) (err error) {
	//fmt.Println("in update method")

	if len(req.World) == 0 {
		err = errors.New("world is empty")
		return
	}

	res.World = make([][]byte, req.P.ImageHeight)
	for row := 0; row < req.P.ImageHeight; row++ {
		res.World[row] = make([]byte, req.P.ImageWidth)
		for col := 0; col < req.P.ImageWidth; col++ {
			res.World[row][col] = req.World[row][col]
		}
	}

	turn := 0
	for turn < req.P.Turns {
		//req.TurnZero <- true
		//fmt.Println("passed true to TZ<-")
		res.World = UpdateBoard(res.World, req.P, req.Events, turn)
		//req.Events <- gol.TurnComplete{CompletedTurns: turn}
		res.CompletedTurns = turn
		fmt.Println("updated res.compTurns=", res.CompletedTurns)
		turn++
	}

	res.AliveCells = calcAliveCells(req.P.ImageHeight, req.P.ImageWidth, res.World)
	res.AliveCellCount = calcAliveCellCount(req.P.ImageHeight, req.P.ImageWidth, res.World)

	//fmt.Println("Updated Response struc: World, Cells, AliveCellCount")
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

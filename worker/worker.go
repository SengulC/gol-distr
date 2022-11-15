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

// UpdateBoard TODO: Update a single iteration
func UpdateBoard(worldIn [][]byte, p gol.Params) [][]byte {
	// worldOut = worldIn
	worldOut := make([][]byte, p.ImageHeight)
	for row := 0; row < p.ImageHeight; row++ {
		worldOut[row] = make([]byte, p.ImageWidth)
		for col := 0; col < p.ImageWidth; col++ {
			worldOut[row][col] = worldIn[row][col]
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

			// if element dead
			if element == 0 {
				if counter == 3 {
					worldOut[row][col] = 255
				}
			} else {
				// if element alive
				if counter < 2 {
					worldOut[row][col] = 0
				} else if counter > 3 {
					worldOut[row][col] = 0
				}
			}
		}
	}

	return worldOut
}

type UpdateOperations struct{}

func (s *UpdateOperations) Update(req gol.Request, res *gol.Response) (err error) {
	fmt.Println("in update method")

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
		fmt.Println("TURN LOOP")
		util.VisualiseMatrix(res.World, req.P.ImageWidth, req.P.ImageHeight)
		res.World = UpdateBoard(res.World, req.P)
		turn++
	}

	var count int
	var cells []util.Cell
	for row := 0; row < req.P.ImageHeight; row++ {
		for col := 0; col < req.P.ImageWidth; col++ {
			if res.World[row][col] == 255 {
				c := util.Cell{X: col, Y: row}
				cells = append(cells, c)
				count++
			}
		}
	}

	res.Cells = cells
	res.AliveCellCount = count

	fmt.Println("Updated Response struc: World, Cells, AliveCellCount")
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
}

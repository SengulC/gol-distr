package gol

import (
	"fmt"
	_ "uk.ac.bris.cs/gameoflife/util"
)

var UpdateBoardHandler = "GOLOperations.UpdateBoard"

type Response struct {
	World [][]byte
}

type Request struct {
	World       [][]byte
	turns       int
	imageHeight int
	imageWidth  int
}

func UpdateBoard(worldIn [][]byte, turns, imageHeight, imageWidth int) [][]byte {
	// worldOut = worldIn
	worldOut := make([][]byte, imageHeight)
	for row := 0; row < imageHeight; row++ {
		worldOut[row] = make([]byte, imageWidth)
		for col := 0; col < imageWidth; col++ {
			worldOut[row][col] = worldIn[row][col]
		}
	}

	turn := 0
	// TODO: Execute all turns of the Game of Life.
	for turn < turns {
		// update board
		// you want worldOut to equal worldIn at the start of every turn.

		for row := 0; row < imageHeight; row++ {
			for col := 0; col < imageWidth; col++ {
				worldOut[row][col] = worldIn[row][col]
			}
		}

		for row := 0; row < imageHeight; row++ {
			for col := 0; col < imageWidth; col++ {
				// CURRENT ELEMENT AND ITS NEIGHBOR COUNT RESET
				element := worldIn[row][col]
				counter := 0

				// iterate through all neighbors of given element
				for dy := -1; dy <= 1; dy++ {
					for dx := -1; dx <= 1; dx++ {
						// creates 3x3 matrix w element as centerpiece, but centerpiece is included as well ofc.
						nRow := (row + dx + imageHeight) % imageHeight
						nCol := (col + dy + imageWidth) % imageWidth
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

		for row := 0; row < imageHeight; row++ {
			for col := 0; col < imageWidth; col++ {
				worldIn[row][col] = worldOut[row][col]
			}
		}

		turn++
	}
}

type GOLOperations struct{}

func (s *GOLOperations) UpdateBoard(req Request, res *Response) (err error) {
	fmt.Println("Got World")
	// for loop
	//	updateBoard()
	res.World = UpdateBoard(Request.World, Request.turns, Request.imageHeight, Request.imageWidth)
	return
}

// RECEIVE DATA FROM FROM DISTRIBUTOR

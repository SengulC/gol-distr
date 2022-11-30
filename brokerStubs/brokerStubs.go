package brokerStub

import (
	"uk.ac.bris.cs/gameoflife/gol"
	"uk.ac.bris.cs/gameoflife/util"
)

var TickerHandler = "UpdateOperations.Ticker"
var SaveHandler = "UpdateOperations.Save"
var PauseHandler = "UpdateOperations.Pause"
var ContinueHandler = "UpdateOperations.Continue"
var UpdateHandler = "WorkerOperations.Update"

type WorkerResponse struct {
	World          [][]byte
	AliveCells     []util.Cell
	CompletedTurns int
	AliveCellCount int
}

type WorkerRequest struct {
	World  [][]byte
	P      gol.Params
	Events chan<- gol.Event
}

type TickerResponse struct {
	World          [][]byte
	AliveCells     []util.Cell
	CompletedTurns int
	AliveCellCount int
}

type TickerRequest struct {
	World  [][]byte
	P      gol.Params
	Events chan<- gol.Event
}

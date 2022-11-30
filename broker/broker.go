package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"time"
	"uk.ac.bris.cs/gameoflife/brokerStubs"
	"uk.ac.bris.cs/gameoflife/gol"
)

//var BrokerSaveHandler = "BrokerOperations.BrokerSave"
//var BrokerContinueHandler = "BrokerOperations.BrokerSave"

var workerServer = flag.String("workerServer", "127.0.0.1:8060", "IP:port string to connect to as server")

type BrokerOperations struct {
	completedTurns int
	aliveCells     int
	currentWorld   [][]byte
	client         *rpc.Client
}

func (b *BrokerOperations) BrokerTicker(req gol.Request, res *gol.Response) (err error) {
	var workerReq = brokerStub.TickerRequest{World: req.World, P: req.P}
	var workerRes = new(brokerStub.TickerResponse)
	b.client.Call(brokerStub.TickerHandler, workerReq, workerRes)
	res.World = workerRes.World
	res.CompletedTurns = workerRes.CompletedTurns
	res.AliveCellCount = workerRes.AliveCellCount
	fmt.Println(workerRes.AliveCellCount, workerRes.CompletedTurns)
	return
}

func (b *BrokerOperations) BrokerSave(req gol.Request, res *gol.Response) (err error) {
	var workerReq = brokerStub.WorkerRequest{World: req.World, P: req.P}
	var workerRes = new(brokerStub.WorkerResponse)
	b.client.Call(brokerStub.SaveHandler, workerReq, workerRes)
	return
}

func (b *BrokerOperations) BrokerPause(req gol.Request, res *gol.Response) (err error) {
	var workerReq = brokerStub.WorkerRequest{World: req.World, P: req.P}
	var workerRes = new(brokerStub.WorkerResponse)
	b.client.Call(brokerStub.PauseHandler, workerReq, workerRes)
	return
}

func (b *BrokerOperations) BrokerContinue(req gol.Request, res *gol.Response) (err error) {
	var workerReq = brokerStub.WorkerRequest{World: req.World, P: req.P}
	var workerRes = new(brokerStub.WorkerResponse)
	b.client.Call(brokerStub.ContinueHandler, workerReq, workerRes)
	return
}

func (b *BrokerOperations) BrokerGOL(req gol.Request, res *gol.Response) (err error) {
	b.client, err = rpc.Dial("tcp", *workerServer)
	if err != nil {
		fmt.Println("ummm error.")
	}
	defer b.client.Close()

	var workerReq = brokerStub.WorkerRequest{World: req.World, P: req.P}
	var workerRes = new(brokerStub.WorkerResponse)
	b.client.Call(brokerStub.UpdateHandler, workerReq, workerRes)

	res.World = workerRes.World
	res.AliveCellCount = workerRes.AliveCellCount
	res.CompletedTurns = workerRes.CompletedTurns
	res.AliveCells = workerRes.AliveCells

	b.currentWorld = workerRes.World
	b.aliveCells = workerRes.AliveCellCount
	b.completedTurns = workerRes.CompletedTurns
	return
}

func main() {
	pAddr := flag.String("port", "8050", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	rpc.Register(&BrokerOperations{})
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	rpc.Accept(listener)
}

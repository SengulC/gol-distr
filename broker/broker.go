package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"time"
	"uk.ac.bris.cs/gameoflife/gol"
)

var UpdateHandler = "UpdateOperations.Update"
var TickerHandler = "UpdateOperations.Ticker"
var SaveHandler = "UpdateOperations.Save"
var PauseHandler = "UpdateOperations.Pause"
var ContinueHandler = "UpdateOperations.Continue"
var workerServer = flag.String("workerServer", "54.243.1.32", "IP:port string to connect to as server")

type BrokerOperations struct {
	completedTurns int
	aliveCells     int
	currentWorld   [][]byte
	//server         *string
}

func (b *BrokerOperations) BrokerGOL(req gol.Request, res *gol.Response) (err error) {
	fmt.Println("in the broker GOL method")
	//b.server = workerServer
	client, err := rpc.Dial("tcp", *workerServer)
	if err != nil {
		fmt.Println("ummm error.")
	}
	defer client.Close()
	brokerReq := gol.Request{World: req.World, P: req.P}
	fmt.Println("REQUEST ON BROKER:", len(brokerReq.World))

	fmt.Println("abt to call update handler")
	var workerRes = new(gol.Response)
	err = client.Call(UpdateHandler, brokerReq, workerRes)
	if err != nil {
		fmt.Println("ERROR!")
	}
	res.World = workerRes.World
	res.AliveCells = workerRes.AliveCells
	res.AliveCellCount = workerRes.AliveCellCount
	res.CompletedTurns = workerRes.CompletedTurns
	fmt.Println("called update handler")
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

	//var server = flag.String("server", "3.91.54.94:8050", "IP:port string to connect to as server")
	//client, _ := rpc.Dial("tcp", *server)
	//defer client.Close()
}

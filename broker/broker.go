package main

import (
	"flag"
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

type BrokerOperations struct {
	completedTurns int
	aliveCells     int
	currentWorld   [][]byte
	server         *string
}

func (b *BrokerOperations) BrokerGOL(req gol.Request, res *gol.Response) (err error) {
	b.server = flag.String("server", "54.243.1.32", "IP:port string to connect to as server")
	client, _ := rpc.Dial("tcp", *b.server)
	defer client.Close()

	client.Call(UpdateHandler, req, res)
	return
}

func main() {
	pAddr := flag.String("port", "8040", "Port to listen on")
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

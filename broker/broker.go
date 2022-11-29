package main

import (
	"flag"
	"math/rand"
	"net"
	"net/rpc"
	"time"
)

func broker() {
	var server = flag.String("server", "3.91.54.94:8050", "IP:port string to connect to as server")
	client, _ := rpc.Dial("tcp", *server)
	defer client.Close()
}

func main() {
	pAddr := flag.String("port", "8050", "Port to listen on")
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	listener, _ := net.Listen("tcp", ":"+*pAddr)
	defer listener.Close()
	rpc.Accept(listener)
}

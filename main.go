package main

import (
	"net/rpc"
	"net"
	"log"
	"./packer"
	"net/rpc/jsonrpc"
)

func main(){
	serverStart()
}


func serverStart() {

	m:= packerMeta()

	arith := packer.Packer{m}
	rpc_server := rpc.NewServer()
	rpc_server.Register(&arith)
	rpc_server.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
	l, e := net.Listen("tcp", ":8222")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		go rpc_server.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}
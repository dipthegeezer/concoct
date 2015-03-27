package main

import (
	"./api"
	"github.com/mitchellh/packer/packer"
	"fmt"
	"net/rpc"
	"net"
	"log"
	"./server"
	"net/rpc/jsonrpc"
	"io/ioutil"
)

func main(){
	go serverStart()

	var err error
	var data []byte
	data, err = ioutil.ReadFile("simple.json")
	if err != nil {
		fmt.Printf("%s", err)
	}

	fmt.Println(data)
	conn, err := net.Dial("tcp", "localhost:8222")

	if err != nil {
		panic(err)
	}
	defer conn.Close()

	args := &server.Args{7, 8, data}
	var reply string

	c := jsonrpc.NewClient(conn)

	for i := 0; i < 1; i++ {

		err = c.Call("Arith.Convert", args, &reply)
		if err != nil {
			log.Fatal("arith error:", err)
		}
		fmt.Printf("Arith: %d*%d=%s\n", args.A, args.B, reply)
	}
}

func serverStart() {
	arith := new(server.Arith)
	rpc_server := rpc.NewServer()
	rpc_server.Register(arith)
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


func simple() {
	m:= packerMeta()
	build := api.Build{m}
	userVars := make(map[string]string)
	tpl, err := packer.ParseTemplateFile("simple.json", userVars)
	if err != nil {
		fmt.Println(fmt.Sprintf("Failed to parse template: %s", err))
	}
	build.Run(tpl)
}

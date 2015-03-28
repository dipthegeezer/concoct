package main

import (
    "fmt"
    "net"
    "log"
    "net/rpc/jsonrpc"
    "io/ioutil"
)


func main(){
    var err error
    var args [1]string
    var data []byte
    data, err = ioutil.ReadFile("/home/dipesh/repo/concoct/simple.json")
    if err != nil {
      fmt.Printf("%s", err)
    }
    args[0] = string(data)

    fmt.Printf("%s\n", args[0])

    conn, err := net.Dial("tcp", "localhost:8222")

    if err != nil {
        panic(err)
    }
    defer conn.Close()

    var reply string

    c := jsonrpc.NewClient(conn)
    for i := 0; i < 1; i++ {

        err = c.Call("Packer.Build", &args, &reply)
        if err != nil {
            log.Fatal("arith error:", err)
        }
        fmt.Printf("Arith: %d*%d=%s\n", i, i, reply)
    }
}

package main

import (
    "fmt"
    "net"
    //"bufio"
    "os"
    //"strings"
    //"time"
)

func main() {
    saddr,err := net.ResolveUDPAddr("udp","10.2.2.1:7070")
    if err != nil {
        fmt.Println("Error udp addr:", err.Error())
        os.Exit(1)
    }

    // further listen to addr
    conn, err := net.DialUDP("udp", nil, saddr)
    if err != nil {
        fmt.Println("Error: ",err)
        return
    }

    fmt.Println("listening adr")

    // reader
    //b := make([]byte, 1024)
    //reader := bufio.NewReader(clCon)
    //msg, err := reader.ReadString(';')
    //if err != nil {
    //    fmt.Println("Error: ",err)
    //    return
    //}
    //println(msg)

    // write
    conn.WriteToUDP([]byte("Hello from client"), saddr)

}

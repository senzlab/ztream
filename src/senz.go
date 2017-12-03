package main

import (
    "fmt"
    "net"
    "bufio"
    "os"
    "strings"
    "time"
)

type Senzie struct {
    name        string
    id          string
	out         chan string
    quit        chan bool
    tik         *time.Ticker
    conn        net.Conn
}

type Senz struct {
    msg         string
    ztype       string
    sender      string
    receiver    string
    attr        map[string]string
    digsig      string
}

// keep connected senzies
var senzies = map[string]*Senzie{}

func main() {
    addr, err := net.ResolveUDPAddr("udp",":7070")
    if err != nil {
        fmt.Println("Error udp addr:", err.Error())
        os.Exit(1)
    }

    // listen for incoming udp packets
    conn, err := net.ListenUDP("udp", addr)
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }

    fmt.Println("Listening on " + "7070")

    for {
        reading(conn)
    }
}

func reading(conn *net.UDPConn) {
    buf := make([]byte, 1024)
    n, raddr, err := conn.ReadFromUDP(buf)
    fmt.Println("Received ", string(buf[0:n]), " from ", raddr)

    if err != nil {
        fmt.Println("Error: ",err)
        return
    }

    // further listen to addr
    clCon, err := net.DialUDP("udp", nil, raddr)
    if err != nil {
        fmt.Println("Error: ",err)
        return
    }

    fmt.Println("listening adr")

    // reader
    //b := make([]byte, 1024)
    reader := bufio.NewReader(clCon)
    msg, err := reader.ReadString(';')
    if err != nil {
        fmt.Println("Error: ",err)
        return
    }
    println(msg)

    // write
    conn.WriteToUDP([]byte("Hello from client"), raddr)
}

func writing(senzie *Senzie)  {
    writer := bufio.NewWriter(senzie.conn)

    // write
    WRITER:
    for {
        select {
        case <- senzie.quit:
            println("quiting/write/tick -- " + senzie.id)
            senzie.tik.Stop()
            break WRITER
        case senz := <-senzie.out:
            println("writing -- " + senzie.id)
            println(senz)
            writer.WriteString(senz + ";")
            writer.Flush()
        case <-senzie.tik.C:
            println("ticking -- " + senzie.id)
            writer.WriteString("TIK;")
            writer.Flush()
        }
    }
}

func parse(msg string)*Senz {
    replacer := strings.NewReplacer(";", "", "\n", "")
    tokens := strings.Split(strings.TrimSpace(replacer.Replace(msg)), " ")
    senz := &Senz {}
    senz.msg = msg
    senz.attr = map[string]string{}

    for i := 0; i < len(tokens); i++ {
        if(i == 0) {
            senz.ztype = tokens[i]
        } else if(i == len(tokens) - 1) {
            // signature at the end
            senz.digsig = tokens[i]
        } else if(strings.HasPrefix(tokens[i], "@")) {
            // receiver @eranga
            senz.receiver = tokens[i][1:]
        } else if(strings.HasPrefix(tokens[i], "^")) {
            // sender ^lakmal
            senz.sender = tokens[i][1:]
        } else if(strings.HasPrefix(tokens[i], "$")) {
            // $key er2232
            key := tokens[i][1:]
            val := tokens[i + 1]
            senz.attr[key] = val
            i ++
        } else if(strings.HasPrefix(tokens[i], "#")) {
            key := tokens[i][1:]
            nxt := tokens[i + 1]

            if(strings.HasPrefix(nxt, "#") || strings.HasPrefix(nxt, "$") ||
                                                strings.HasPrefix(nxt, "@")) {
                // #lat #lon
                // #lat @eranga
                // #lat $key 32eewew
                senz.attr[key] = ""
            } else {
                // #lat 3.2323 #lon 5.3434
                senz.attr[key] = nxt
                i ++
            }
        }
    }

    return senz
}

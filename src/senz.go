package main

import (
    "fmt"
    "net"
    "os"
    "strings"
)

type Senz struct {
    msg         string
    ztype       string
    sender      string
    receiver    string
    attr        map[string]string
    digsig      string
}

// keep connected senzies
var senzies = map[string]*net.UDPAddr{}
var streams = map[int]*net.UDPAddr{}

func main() {
    addr, err := net.ResolveUDPAddr("udp", ":" + config.switchPort)
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

    fmt.Println("Listening on " + config.switchPort)

    for {
        reading(conn)
    }
}

func reading(conn *net.UDPConn) {
    buf := make([]byte, 1024)
    n, fAdr, err := conn.ReadFromUDP(buf)

    if err != nil {
        fmt.Println("Error: ",err)
        return
    }

    msg := string(buf[0:n])
    fmt.Println("Received ", msg, " from ", fAdr)

    if(strings.HasPrefix(msg, "DATA")) {
        // handshake msg
        senz := parse(msg)
        if(senz.attr["STREAM"] == "ON") {
            // DATA #STREAM ON #TO eranga ^lakmal digisg
            from := senz.sender
            to := senz.attr["TO"]
            senzies[from] = fAdr

            // check to exists and add to streams
            if tAdr, ok := senzies[to]; ok {
                streams[fAdr.Port] = tAdr
                streams[tAdr.Port] = fAdr
            }
        } else if(senz.attr["STREAM"] == "OFF") {
            // DATA #STREAM OFF #TO eranga ^lakmal digisg
            from := senz.sender
            delete(senzies, from)
            delete(streams, fAdr.Port)
        }
    } else {
        // this is stream forward
        conn.WriteToUDP(buf[0:n], streams[fAdr.Port])
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

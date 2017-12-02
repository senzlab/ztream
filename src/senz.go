package main

import (
    "fmt"
    "net"
    "bufio"
    "os"
    "strings"
    "time"
    "gopkg.in/mgo.v2"
)

type Senzie struct {
    name        string
	outgoing    chan string
    ticking     chan string
    quit        chan bool
	reader      *bufio.Reader
	writer      *bufio.Writer
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
var keyStore = &KeyStore{}

func main() {
    // listen for incoming conns
    l, err := net.Listen("tcp", ":" + config.switchPort)
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }

    // close listern on app closes
    defer l.Close()

    fmt.Println("Listening on " + config.switchPort)

    // db setup
    session, err:= mgo.Dial(config.mongoHost)
    if err != nil {
        fmt.Println("Error connecting mongo: ", err.Error())
        os.Exit(1)
    }

    // close session on app closes
    defer session.Close()

    session.SetMode(mgo.Monotonic, true)
    keyStore.session = session

    for {
        // handle new connections 
        conn, err := l.Accept()
        if err != nil {
            fmt.Println("Error accepting: ", err.Error())
            os.Exit(1)
        }

        // new senzie
        senzie := &Senzie {
            outgoing: make(chan string),
            ticking : make(chan string),
            quit: make(chan bool),
            reader: bufio.NewReader(conn),
            writer: bufio.NewWriter(conn),
        }

        go reading(senzie)
        go writing(senzie)
    }
}

func reading(senzie *Senzie) {
    // read senz
    for {
        msg, err := senzie.reader.ReadString(';')
        if err != nil {
            fmt.Println("Error reading: ", err.Error())
            break
        }

        // parse senz
        var senz = parse(msg)

        if(senz.receiver == config.switchName) {
            if(senz.ztype == "SHARE") {
                // this is shareing pub key(registration)
                println("SHARE pubKey to switch")

                // TODO save pubkey in db

                // senzie registered
                senzie.name = senz.sender
                senzies[senzie.name] = senzie

                // start ticking
                go ticking(senzie)
            } else if(senz.ztype == "GET") {
                // this is requesting pub key of other senzie
            }
        } else {
            // senz for another senzie
            println("SENZ for senzie")

            // forwared senz
            var senzie = senzies[senz.receiver]
            senzie.outgoing <- senz.digsig
        }
    }

    // senzie exists
    // quit all routeins
    senzie.quit <- true
}

func ticking(senzie *Senzie) {
    // ping
    for {
        select {
        case <- senzie.quit:
            println("quiting -- ")
            break
        default:
            <-time.After(120 * time.Second)
            senzie.ticking <- "TIK"
        }
    }
}

func writing(senzie *Senzie)  {
    // write
    for {
        select {
        case <- senzie.quit:
            println("quiting -- ")
            break
        case senz := <-senzie.outgoing:
            println("writing -- ")
            senzie.writer.WriteString(senz + ";")
            senzie.writer.Flush()
        case tick := <-senzie.ticking:
            println("ticking -- ")
            senzie.writer.WriteString(tick + ";")
            senzie.writer.Flush()
        }
    }
}

func parse(msg string)*Senz {
    var replacer = strings.NewReplacer(";", "", "\n", "")
    var tokens = strings.Split(strings.TrimSpace(replacer.Replace(msg)), " ")
    var senz = &Senz {}
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
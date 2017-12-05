package main

import (
    "net"
    "strconv"
    "time"
    //"fmt"
)

func main() {
	addr,_ := net.ResolveUDPAddr("udp","203.94.69.162:7070")


    t3 := int64(time.Now().UnixNano())
    println(addrToSr(addr))
    t4 := int64(time.Now().UnixNano())
    println(t4 - t3)

    t1 := int64(time.Now().UnixNano())
    println(addrToSt(addr))
    t2 := int64(time.Now().UnixNano())
    println(t2 - t1)

    t5 := int64(time.Now().UnixNano())
    println(addr.String())
    t6 := int64(time.Now().UnixNano())
    println(t6 - t5)
}

func addrToSt(addr *net.UDPAddr)string {
    return addr.IP.String() + strconv.Itoa(addr.Port)
}

func addrToSr(addr *net.UDPAddr)string {
    return string(addr.IP.To4()) + strconv.Itoa(addr.Port)
}

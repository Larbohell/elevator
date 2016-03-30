package network

import "net"
import "fmt"
import "time"

func CheckError(err error) {
    if err  != nil {
        fmt.Println("Error: " , err)
    }
}

func Udp_broadcast(message string){
	BroadcastAddr, err := net.ResolveUDPAddr("udp4", "129.241.187.255:20011")
	CheckError(err)

	BroadcastConn, err := net.DialUDP("udp4", nil, BroadcastAddr)
	CheckError(err)

	_ , err = BroadcastConn.Write([]byte(message))
	CheckError(err)

	BroadcastConn.Close()
}

func Udp_listen(port string) string {
	ListenerAddr, err := net.ResolveUDPAddr("udp4", port)
	CheckError(err)

	listenConn, err := net.ListenUDP("udp4", ListenerAddr)
	CheckError(err)

	var listenBuffer[1024]byte

	deadlineTime := time.Now().Add(time.Second)
	_, _, err = listenConn.ReadFromUDP(listenBuffer[0:])
	DeadlineError :=  listenConn.SetReadDeadline(deadlineTime)

	CheckError(err)


	if (DeadlineError != nil){
		listenConn.Close()
		return ""
	}

	listenConn.Close()

	return string(listenBuffer[0:1024])
}

//UDP server at 129.241.287.23:52038

package network

import (
	"fmt"
	"net"
	"strconv"
	"encoding/json"
	"elevator_type"
)

//UDP server at 129.241.187.23:52038

var laddr *net.UDPAddr //Local address
var baddr *net.UDPAddr //Broadcast address

type Udp_message struct {
	Raddr  string //if receiving raddr=senders address, if sending raddr should be set to "broadcast" or an ip:port
	Data   string //TODO: implement another encoding, strings are meh
	Length int    //length of received data, in #bytes // N/A for sending
}

//func Udp_init(localListenPort, broadcastListenPort, message_size int, send_ch, receive_ch chan Udp_message) (err error) {
func Udp_init(localListenPort, broadcastListenPort, message_size int, send_ch, receive_ch chan elevator_type.Elevator) (err error) {
	//Generating broadcast address
	baddr, err = net.ResolveUDPAddr("udp4", "255.255.255.255:"+strconv.Itoa(broadcastListenPort))
	if err != nil {
		return err
	}

	//Generating localaddress
	tempConn, err := net.DialUDP("udp4", nil, baddr)
	defer tempConn.Close()
	tempAddr := tempConn.LocalAddr()
	laddr, err = net.ResolveUDPAddr("udp4", tempAddr.String())
	laddr.Port = localListenPort

	//Creating local listening connections
	localListenConn, err := net.ListenUDP("udp4", laddr)
	if err != nil {
		return err
	}

	//Creating listener on broadcast connection
	broadcastListenConn, err := net.ListenUDP("udp", baddr)
	if err != nil {
		localListenConn.Close()
		return err
	}

	go udp_receive_server(localListenConn, broadcastListenConn, message_size, receive_ch)
	go udp_transmit_server(localListenConn, broadcastListenConn, send_ch)

	//	fmt.Printf("Generating local address: \t Network(): %s \t String(): %s \n", laddr.Network(), laddr.String())
	//	fmt.Printf("Generating broadcast address: \t Network(): %s \t String(): %s \n", baddr.Network(), baddr.String())
	return err
}

func udp_transmit_server(lconn, bconn *net.UDPConn, send_ch chan elevator_type.Elevator) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("ERROR in udp_transmit_server: %s \n Closing connection.", r)
			lconn.Close()
			bconn.Close()
		}
	}()

	//var err error
	//var n int
	baddr, _ := net.ResolveUDPAddr("udp", "129.241.187.255:30000")
	lconn, _ = net.DialUDP("udp", nil ,baddr) 
	var msg elevator_type.Elevator

	for {
		//		fmt.Printf("udp_transmit_server: waiting on new value on Global_Send_ch \n")
		msg = <-send_ch
		//		fmt.Printf("Writing %s \n", msg.Data)
		
		/*
		if msg.Raddr == "broadcast" {
			n, err = lconn.WriteToUDP([]byte(msg.Data), baddr)
		} else {
			raddr, err := net.ResolveUDPAddr("udp", msg.Raddr)
			if err != nil {
				fmt.Printf("Error: udp_transmit_server: could not resolve raddr\n")
				panic(err)
			}
			n, err = lconn.WriteToUDP([]byte(msg.Data), raddr)
		}
		*/

		buf, err := json.Marshal(msg)
		_, err = lconn.Write(buf)

		if err != nil {
			fmt.Printf("Error: udp_transmit_server: writing\n")
			panic(err)
		}
		//		fmt.Printf("udp_transmit_server: Sent %s to %s \n", msg.Data, msg.Raddr)
	}
}

func udp_receive_server(lconn, bconn *net.UDPConn, message_size int, receive_ch chan elevator_type.Elevator) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("ERROR in udp_receive_server: %s \n Closing connection.", r)
			lconn.Close()
			bconn.Close()
		}
	}()

	bconn_rcv_ch := make(chan elevator_type.Elevator)
	lconn_rcv_ch := make(chan elevator_type.Elevator)

	go udp_connection_reader(lconn, message_size, lconn_rcv_ch)
	go udp_connection_reader(bconn, message_size, bconn_rcv_ch)

	for {
		select {

		case buf := <-bconn_rcv_ch:
			receive_ch <- buf

		case buf := <-lconn_rcv_ch:
			receive_ch <- buf
		}
	}
}

func udp_connection_reader(conn *net.UDPConn, message_size int, rcv_ch chan elevator_type.Elevator) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("ERROR in udp_connection_reader: %s \n Closing connection.", r)
			conn.Close()
		}
	}()

	var received_message elevator_type.Elevator
	raddr, _ := net.ResolveUDPAddr("udp", ":30000")
	conn, _ = net.ListenUDP("udp", raddr)

	for {
		buffer := make([]byte, 4098)
		message_length, _, _ := conn.ReadFromUDP(buffer)
		err := json.Unmarshal(buffer[:message_length], &received_message)
		
		
		//		fmt.Printf("udp_connection_reader: Waiting on data from UDPConn\n")
		//n, raddr, err := conn.ReadFromUDP(buf)
		//		fmt.Printf("udp_connection_reader: Received %s from %s \n", string(buf), raddr.String())
		
		if err != nil || message_length < 0 {
			fmt.Printf("Error: udp_connection_reader: reading\n")
			panic(err)
		}
		//rcv_ch <- Udp_message{Raddr: raddr.String(), Data: string(buf), Length: n}
		rcv_ch <- received_message
	}
}
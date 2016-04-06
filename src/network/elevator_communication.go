package network
import "strconv"

// This makes the elevator stuck in udp.go
// Must probably be ran in separate thread from main execution


func Send_floor(floor int){
	send_ch := make(chan Udp_message)
	receive_ch := make(chan Udp_message)

	_ = Udp_init(20011, 30000, 1024, send_ch, receive_ch)

	send_msg := Udp_message{Raddr: "129.241.187.24:20011", Data: strconv.Itoa(floor)}
	//send_msg := network.Udp_message{Raddr: "broadcast", Data: "hello world"}
	send_ch <- send_msg
}

func Receive_floor() int {
	send_ch := make(chan Udp_message)
	receive_ch := make(chan Udp_message)

	_ = Udp_init(20011, 30000, 1024, send_ch, receive_ch)
	
	var receive_msg Udp_message

	receive_msg = <- receive_ch
	receivedFloor, _ := strconv.Atoi(receive_msg.Data[10:11])
	return receivedFloor
}
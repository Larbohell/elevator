package network
//import "strconv"
import "elevator_type"
import "fmt"

// This makes the elevator stuck in udp.go
// Must probably be ran in separate thread from main execution

/*
func Send_Floor(Floor int){
	send_ch := make(chan Udp_message)
	receive_ch := make(chan Udp_message)

	_ = Udp_init(20011, 30000, 1024, send_ch, receive_ch)

	send_msg := Udp_message{Raddr: "129.241.187.24:20011", Data: strconv.Itoa(Floor)}
	//send_msg := network.Udp_message{Raddr: "broadcast", Data: "hello world"}
	send_ch <- send_msg
}

func Receive_Floor() int {
	send_ch := make(chan Udp_message)
	receive_ch := make(chan Udp_message)

	_ = Udp_init(20011, 30000, 1024, send_ch, receive_ch)
	
	var receive_msg Udp_message

	receive_msg = <- receive_ch
	receivedFloor, _ := strconv.Atoi(receive_msg.Data[10:11])
	return receivedFloor
}
*/

func Send_elevator_struct(){
	for {
		send_ch := make(chan elevator_type.Elevator)
		receive_ch := make(chan elevator_type.Elevator)
		_ = Udp_init(20011, 30000, 1024, send_ch, receive_ch)	

		var array [4][3]int = [4][3]int{{0, 0, 0}, {1, 0, 1}, {0, 1, 1}, {1, 1, 1}}

		send_msg := elevator_type.Elevator{
			Floor: 1,
			Dir: elevator_type.Up,
			Requests: array,
			Behaviour: elevator_type.EB_DoorOpen,
		} 
		//var a [3][3]int = [3][3]int{{1,2,3},{4,5,6},{7,8,9}}
		send_ch <- send_msg
	}
}

func Receive_elevator_struct(){
	send_ch := make(chan elevator_type.Elevator)
	receive_ch := make(chan elevator_type.Elevator)

	_ = Udp_init(20011, 30000, 1024, send_ch, receive_ch)
	
	var receive_msg elevator_type.Elevator

	receive_msg = <- receive_ch
	fmt.Println("Floor: ", receive_msg.Floor, "Dir: ", receive_msg.Dir, "behavior: ", receive_msg.Behaviour)
}
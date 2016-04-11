package main

import "driver"
import "elevator"
import "elevator_type"
//import "network"

//import "fmt"
//import "strconv"

//"129.241.187.156" = workspace 9
//"129.241.187.159" = workspace 11

func main() {
		driver.Elevator_init()


		/*
		if (driver.Elevator_get_floor_sensor_signal() == -1){
			elevator.Fsm_onInitBetweenFloors()
		} else {
			elevator.Fsm_onInitOnFloor()
		}

		var prevFloor int = driver.Elevator_get_floor_sensor_signal() //default value
		var currentFloor int
		var previous_button_value[elevator_type.N_FLOORS][elevator_type.N_BUTTONS] bool
		*/


		//go checkTimer()
		/*
		//On button pressed
		for {
				for floor := 0; floor < elevator_type.N_FLOORS; floor++ {
					for button := 0; button < elevator_type.N_BUTTONS; button++ {

						var button_value bool = driver.Elevator_is_button_pushed(driver.Button(button), floor)
						if (button_value && button_value != previous_button_value[floor][button]){
							elevator.Fsm_onRequestButtonPress(floor, driver.Button(button))
						}
						previous_button_value[floor][button] = button_value
					}
				}
			

		
		
		
			//On floor arrival
			currentFloor = driver.Elevator_get_floor_sensor_signal()

			if currentFloor != -1 && currentFloor != prevFloor {
				elevator.Fsm_onFloorArrival(currentFloor)
			}
			prevFloor = currentFloor
		
		}
		*/


	//Testing UDP module
		/*
	send_ch := make(chan network.Udp_message)
	receive_ch := make(chan network.Udp_message)

	_ = network.Udp_init(20011, 30000, 1024, send_ch, receive_ch)

	var eCurrentFloor int = 2
	var data string = strconv.Itoa(eCurrentFloor)
	send_msg := network.Udp_message{Raddr: "129.241.187.24:20011", Data: data}
	//send_msg := network.Udp_message{Raddr: "broadcast", Data: "hello world"}
	
	var receive_msg network.Udp_message

	for {
		send_ch <- send_msg
		receive_msg = <- receive_ch
		receivedFloor, _ := strconv.Atoi(receive_msg.Data[10:11])
		//11receivedFloor, _ := strconv.Atoi("2")
		receivedFloor = receivedFloor + 5
		fmt.Println(receivedFloor)
		//fmt.Println(receive_msg.Data[10:11])
	}
	*/
}

/*
func checkTimer(){
	for {
		if elevator.Timer_timedout(){
			elevator.Fsm_onDoorTimeout()
			elevator.Timer_stop()
		}
	}
}*/

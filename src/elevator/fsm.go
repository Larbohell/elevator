package elevator

import "driver"
import "network"
import "fmt"
import "elevator_type"

var elevator elevator_type.Elevator

func Fsm_updateAllLights(elevator elevator_type.Elevator){
	for Floor := 0; Floor < elevator_type.N_FLOORS; Floor++{
		for btn := 0; btn < elevator_type.N_BUTTONS; btn++{
			driver.Elevator_set_button_lamp(driver.Button(btn), Floor, elevator.Requests[Floor][btn])
		}
	}
}

func Fsm_onInitOnFloor(){
	// Handle saved Requests
	driver.Elevator_set_motor_direction(driver.MOTOR_DIRECTION_STOP)
	elevator.Dir = elevator_type.Stop
	elevator.Floor = driver.Elevator_get_floor_sensor_signal()
	elevator.Behaviour = elevator_type.EB_Idle
	Fsm_updateAllLights(elevator)
	Print_status(elevator);
}

func Fsm_onInitBetweenFloors(){
	driver.Elevator_set_motor_direction(driver.MOTOR_DIRECTION_DOWN)
	elevator.Dir = elevator_type.Down
	elevator.Floor = -1
	elevator.Behaviour = elevator_type.EB_Moving
	Print_status(elevator);
}

func Fsm_onFloorArrival(newFloor int) {
	elevator.Floor = newFloor
	driver.Elevator_set_floor_indicator(newFloor)

	switch (elevator.Behaviour){
		case elevator_type.EB_Moving:
			if (Requests_shouldStop(elevator)){
				driver.Elevator_set_motor_direction(driver.MOTOR_DIRECTION_STOP)
				driver.Elevator_set_door_open_lamp(1)
				elevator = Requests_clearAtCurrentFloor(elevator)
				// Notify other elevators on the network
				elevator.Behaviour = elevator_type.EB_DoorOpen
				go Timer_start()
				Fsm_updateAllLights(elevator)
			}
	}
	Print_status(elevator);
}

func Fsm_onRequestButtonPress(button_floor int, button driver.Button){
	///// TEST SHIT ////////
	
	var msg elevator_type.Message
	msg.FromMaster = false
	msg.ElevatorInfo = elevator
	msg.ButtonInfo = int(button)

	listenUDPchan := make(chan elevator_type.Message, 1)

	go network.SendUdpMessage(msg)
	go network.ReceiveUdpMessage(listenUDPchan)

	message_received := <- listenUDPchan
	fmt.Println("Floor: ", message_received.ElevatorInfo.Floor)
	
	////////////////////////////////////////////////
	//network.Send_floor(button_floor)
	// ELEVATOR STUCK HERE
	///////////////////////////////////////////////

	//fmt.Println("Here")
	//received_Floor := network.Receive_Floor()
	//fmt.Println("There")
	//fmt.Println("Received Floor: %d\n", received_Floor)

	/*
	if (external_button){
		switch(elevator_role){
		case MASTER:
			elevator_ID = calc_best_suited_elevator(button_Floor int, button driver.Button)
			if (elevator_ID == my_elevator_ID){
				Fsm_handle_order()
			}
			else {
				send_order(elevator_ID, Floor)
			}
		case SLAVE:
			inform_master(button_Floor int, button driver.Button);
	}
	///////// Move code below to handle_order() function or something similar
	*/
	
	switch (elevator.Behaviour){
	case elevator_type.EB_DoorOpen:
		if (elevator.Floor == button_floor){
			go Timer_start()
		} else {
			elevator.Requests[button_floor][button] = 1
		}
	case elevator_type.EB_Moving:
		elevator.Requests[button_floor][button] = 1
	case elevator_type.EB_Idle:
		elevator.Requests[button_floor][button] = 1
		elevator.Dir = Requests_chooseDirection(elevator)

		if (elevator.Dir == elevator_type.Stop){
			driver.Elevator_set_door_open_lamp(1)
			elevator = Requests_clearAtCurrentFloor(elevator)
			elevator.Behaviour = elevator_type.EB_DoorOpen
			go Timer_start()

		} else {
			driver.Elevator_set_motor_direction(driver.Motor_direction(elevator.Dir))
			elevator.Behaviour = elevator_type.EB_Moving
		}
	}
	Fsm_updateAllLights(elevator)
	Print_status(elevator);
}

/*
func send_button_press(button_Floor int, button driver.Button){
	switch (elevator_role){
		case MASTER:
			elevator_ID = calc_best_suited_elevator()
			if (elevator_ID == my_elevator_ID){
				handle_order()
			}
			else {
				send_order(elevator_ID, Floor)
			}
		
		case SLAVE:
			inform_master(button_Floor int, button driver.Button)
	}
}

func send_order(elevator_Id, Floor int){
	// UDP shit
}

func calc_best_suited_elevator(){
	// COST FUNCTION
	// Do math
	// Simple: Difference between current Floor and destination Floor
}

func Fsm_onOrderReceived(){

}

*/

func Fsm_onDoorTimeout(){
	switch(elevator.Behaviour){
	case elevator_type.EB_DoorOpen:
		elevator.Dir = Requests_chooseDirection(elevator)
		driver.Elevator_set_door_open_lamp(0)
		driver.Elevator_set_motor_direction(driver.Motor_direction(elevator.Dir))

		if (elevator.Dir == elevator_type.Stop){
			elevator.Behaviour = elevator_type.EB_Idle
		} else {
			elevator.Behaviour = elevator_type.EB_Moving
		}

	}
	Print_status(elevator);
}

func Driver(){


	for {
		select {
		case movingDirection := <- setMovingDirectionChan:
		Elevator_set_motor_direction(movingDirection)

		}
	}
}





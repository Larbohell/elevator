package elevator

import "driver"

var elevator Elevator

func Fsm_updateAllLights(elevator Elevator){
	for floor := 0; floor < N_FLOORS; floor++{
		for btn := 0; btn < N_BUTTONS; btn++{
			driver.Elevator_set_button_lamp(driver.Button(btn), floor, elevator.requests[floor][btn])
		}
	}
}

func Fsm_onInitOnFloor(){
	// Handle saved requests
	driver.Elevator_set_motor_direction(driver.MOTOR_DIRECTION_STOP)
	elevator.dir = Stop
	elevator.floor = driver.Elevator_get_floor_sensor_signal()
	elevator.behaviour = EB_Idle
	Fsm_updateAllLights(elevator)
	Print_status(elevator);
}

func Fsm_onInitBetweenFloors(){
	driver.Elevator_set_motor_direction(driver.MOTOR_DIRECTION_DOWN)
	elevator.dir = Down
	elevator.floor = -1
	elevator.behaviour = EB_Moving
	Print_status(elevator);
}

func Fsm_onFloorArrival(newFloor int) {
	elevator.floor = newFloor
	driver.Elevator_set_floor_indicator(newFloor)

	switch (elevator.behaviour){
		case EB_Moving:
			if (Requests_shouldStop(elevator)){
				driver.Elevator_set_motor_direction(driver.MOTOR_DIRECTION_STOP)
				driver.Elevator_set_door_open_lamp(1)
				elevator = Requests_clearAtCurrentFloor(elevator)
				// Notify other elevators on the network
				elevator.behaviour = EB_DoorOpen
				go Timer_start()
				Fsm_updateAllLights(elevator)
			}
	}
	Print_status(elevator);
}

func Fsm_onRequestButtonPress(button_floor int, button driver.Button){
	if (external_button){
		switch(elevator_role){
		case MASTER:
			elevator_ID = calc_best_suited_elevator(button_floor int, button driver.Button)
			if (elevator_ID == my_elevator_ID){
				Fsm_handle_order()
			}
			else {
				send_order(elevator_ID, floor)
			}
		case SLAVE:
			inform_master(button_floor int, button driver.Button);
	}
	///////// Move code below to handle_order() function or something similar
	
	
	switch (elevator.behaviour){
	case EB_DoorOpen:
		if (elevator.floor == button_floor){
			go Timer_start()
		} else {
			elevator.requests[button_floor][button] = 1
		}
	case EB_Moving:
		elevator.requests[button_floor][button] = 1
	case EB_Idle:
		elevator.requests[button_floor][button] = 1
		elevator.dir = Requests_chooseDirection(elevator)

		if (elevator.dir == Stop){
			driver.Elevator_set_door_open_lamp(1)
			elevator = Requests_clearAtCurrentFloor(elevator)
			elevator.behaviour = EB_DoorOpen
			go Timer_start()

		} else {
			driver.Elevator_set_motor_direction(driver.Motor_direction(elevator.dir))
			elevator.behaviour = EB_Moving
		}
	}
	Fsm_updateAllLights(elevator)
	Print_status(elevator);
}

func send_button_press(button_floor int, button driver.Button){
	switch (elevator_role){
		case MASTER:
			elevator_ID = calc_best_suited_elevator()
			if (elevator_ID == my_elevator_ID){
				handle_order()
			}
			else {
				send_order(elevator_ID, floor)
			}
		
		case SLAVE:
			inform_master(button_floor int, button driver.Button)
	}
}

func send_order(elevator_Id, floor int){
	// UDP shit
}

func calc_best_suited_elevator(){
	// COST FUNCTION
	// Do math
	// Simple: Difference between current floor and destination floor
}

func Fsm_onOrderReceived(){

}

func Fsm_onDoorTimeout(){
	switch(elevator.behaviour){
	case EB_DoorOpen:
		elevator.dir = Requests_chooseDirection(elevator)
		driver.Elevator_set_door_open_lamp(0)
		driver.Elevator_set_motor_direction(driver.Motor_direction(elevator.dir))

		if (elevator.dir == Stop){
			elevator.behaviour = EB_Idle
		} else {
			elevator.behaviour = EB_Moving
		}

	}
	Print_status(elevator);
}







package main
import "driver"
import "elevator"
//import "fmt"

func main() {
	driver.Elevator_init()
	
	if (driver.Elevator_get_floor_sensor_signal() == -1){
		elevator.Fsm_onInitBetweenFloors()
	} else {
		elevator.Fsm_onInitOnFloor()
	}

	var prevFloor int = driver.Elevator_get_floor_sensor_signal() //default value
	var currentFloor int
	var previous_button_value[elevator.N_FLOORS][elevator.N_BUTTONS] bool

	//go checkTimer()

	//On button pressed
	for {
			for floor := 0; floor < elevator.N_FLOORS; floor++ {
				for button := 0; button < elevator.N_BUTTONS; button++ {

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
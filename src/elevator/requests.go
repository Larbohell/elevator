package elevator
import "driver"

func Requests_shouldStop(elevator Elevator) bool {
	// Returns true if anyone wants off or on (in the direction of travel), or if a request
	// has been cleared while the elevator was on its way

	switch (elevator.dir){
	case Down:
		if elevator.requests[elevator.floor][driver.BUTTON_OUTSIDE_DOWN] == 1 ||
		elevator.requests[elevator.floor][driver.BUTTON_INSIDE_COMMAND] == 1|| 
		!requests_below(elevator){
			return true
		}
	case Up:
		if elevator.requests[elevator.floor][driver.BUTTON_OUTSIDE_UP] == 1||
		elevator.requests[elevator.floor][driver.BUTTON_INSIDE_COMMAND] == 1|| 
		!requests_above(elevator){
			return true
		}
	case Stop:
		return true
	}
	return false
}

func Requests_clearAtCurrentFloor(elevator Elevator)Elevator{
	for btn := 0; btn < N_BUTTONS; btn++{
		elevator.requests[elevator.floor][btn] = 0
	}
	return elevator
}

func requests_above(elevator Elevator) bool {
	for floor := elevator.floor + 1; floor < N_FLOORS; floor++{
		for button := 0; button < N_BUTTONS; button++{
			if elevator.requests[floor][button] == 1{
				return true
			}
		}
	}
	return false
}

func requests_below(elevator Elevator) bool {
	for floor := 0; floor < elevator.floor; floor++{
		for button := 0; button < N_BUTTONS; button++{
			if elevator.requests[floor][button] == 1{
				return true
			}
		}
	}
	return false
}

func nearest_request_direction(elevator Elevator){

}

func Requests_chooseDirection(elevator Elevator) Dir {
	switch(elevator.dir){
	case Up:
		if requests_above(elevator) {
			return Up
		} else if requests_below(elevator) {
			return Down
		} else {
			return Stop
		}

	case Down:
		if requests_below(elevator) {
			return Down
		} else if requests_above(elevator) {
			return Up
		} else {
			return Stop
		}

	case Stop:
		if requests_below(elevator) {
			return Down
		} else if requests_above(elevator) {
			return Up
		} else {
			return Stop
		}
		//return nearest_request_direction()
	}



	return Up
}
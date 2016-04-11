package elevator
import "driver"
import "elevator_type"

func Requests_shouldStop(elevator elevator_type.Elevator) bool {
	// Returns true if anyone wants off or on (in the Direction of travel), or if a request
	// has been cleared while the elevator was on its way

	switch (elevator.Dir){
	case elevator_type.Down:
		if elevator.Requests[elevator.Floor][driver.BUTTON_OUTSIDE_DOWN] == 1 ||
		elevator.Requests[elevator.Floor][driver.BUTTON_INSIDE_COMMAND] == 1|| 
		!Requests_below(elevator){
			return true
		}
	case elevator_type.Up:
		if elevator.Requests[elevator.Floor][driver.BUTTON_OUTSIDE_UP] == 1||
		elevator.Requests[elevator.Floor][driver.BUTTON_INSIDE_COMMAND] == 1|| 
		!Requests_above(elevator){
			return true
		}
	case elevator_type.Stop:
		return true
	}
	return false
}

func Requests_clearAtCurrentFloor(elevator elevator_type.Elevator) elevator_type.Elevator{
	for btn := 0; btn < elevator_type.N_BUTTONS; btn++{
		elevator.Requests[elevator.Floor][btn] = 0
	}
	return elevator
}

func Requests_above(elevator elevator_type.Elevator) bool {
	for Floor := elevator.Floor + 1; Floor < elevator_type.N_FLOORS; Floor++{
		for button := 0; button < elevator_type.N_BUTTONS; button++{
			if elevator.Requests[Floor][button] == 1{
				return true
			}
		}
	}
	return false
}

func Requests_below(elevator elevator_type.Elevator) bool {
	for Floor := 0; Floor < elevator.Floor; Floor++{
		for button := 0; button < elevator_type.N_BUTTONS; button++{
			if elevator.Requests[Floor][button] == 1{
				return true
			}
		}
	}
	return false
}

func nearest_request_Direction(elevator elevator_type.Elevator){

}

func Requests_chooseDirection(elevator elevator_type.Elevator) elevator_type.Dir {
	switch(elevator.Dir){
	case elevator_type.Up:
		if Requests_above(elevator) {
			return elevator_type.Up
		} else if Requests_below(elevator) {
			return elevator_type.Down
		} else {
			return elevator_type.Stop
		}

	case elevator_type.Down:
		if Requests_below(elevator) {
			return elevator_type.Down
		} else if Requests_above(elevator) {
			return elevator_type.Up
		} else {
			return elevator_type.Stop
		}

	case elevator_type.Stop:
		if Requests_below(elevator) {
			return elevator_type.Down
		} else if Requests_above(elevator) {
			return elevator_type.Up
		} else {
			return elevator_type.Stop
		}
		//return nearest_request_Direction()
	}



	return elevator_type.Up
}
package elevator

import "driver"
import . "elevator_type"
import . "errorHandler"

func Requests_shouldStop(elevator ElevatorInfo) bool {
	// Returns true if anyone wants off or on (in the Direction of travel), or if a request
	// has been cleared while the elevator was on its way

	switch elevator.Direction {
	case Down:
		if elevator.Requests[elevator.CurrentFloor][driver.BUTTON_OUTSIDE_DOWN] == 1 ||
			elevator.Requests[elevator.CurrentFloor][driver.BUTTON_INSIDE_COMMAND] == 1 ||
			!Requests_below(elevator) {
			return true
		}
	case Up:
		if elevator.Requests[elevator.CurrentFloor][driver.BUTTON_OUTSIDE_UP] == 1 ||
			elevator.Requests[elevator.CurrentFloor][driver.BUTTON_INSIDE_COMMAND] == 1 ||
			!Requests_above(elevator) {
			return true
		}
	case Stop:
		errorChannel <- "Function 'Requests_shouldStop': Elevator.Direction is already 'Stop'."
		return true
	}
	return false
}

func Requests_clearAtCurrentFloor(elevator ElevatorInfo) ElevatorInfo {
	for btn := 0; btn < N_BUTTONS; btn++ {
		elevator.Requests[elevator.CurrentFloor][btn] = 0
	}
	return elevator
}

func Requests_above(elevator ElevatorInfo) bool {
	for floor := elevator.CurrentFloor + 1; floor < N_FLOORS; floor++ {
		for button := 0; button < N_BUTTONS; button++ {
			if elevator.Requests[floor][button] == 1 {
				return true
			}
		}
	}
	return false
}

func Requests_below(elevator ElevatorInfo) bool {
	for floor := 0; floor < elevator.CurrentFloor; floor++ {
		for button := 0; button < N_BUTTONS; button++ {
			if elevator.Requests[floor][button] == 1 {
				return true
			}
		}
	}
	return false
}

func nearest_request_Direction(elevator ElevatorInfo) {

}

func Requests_chooseDirection(elevator ElevatorInfo) Dir {
	switch elevator.Direction {
	case Up:
		if Requests_above(elevator) {
			return Up
		} else if Requests_below(elevator) {
			return Down
		} else {
			return Stop
		}

	case Down:
		if Requests_below(elevator) {
			return Down
		} else if Requests_above(elevator) {
			return Up
		} else {
			return Stop
		}

	case Stop:
		if Requests_below(elevator) {
			return Down
		} else if Requests_above(elevator) {
			return Up
		} else {
			return Stop
		}
	}
}

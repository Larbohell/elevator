package orderHandler

import "elevator_type"

func ShouldStop(elevator elevator_type.ElevatorInfo) bool {
	switch elevator.Direction {
	case elevator_type.Down:
		if elevator.Requests[elevator.CurrentFloor][elevator_type.BUTTON_OUTSIDE_DOWN] == 1 ||
			elevator.Requests[elevator.CurrentFloor][elevator_type.BUTTON_INSIDE_COMMAND] == 1 ||
			!requests_below(elevator) {
			return true
		}
	case elevator_type.Up:
		if elevator.Requests[elevator.CurrentFloor][elevator_type.BUTTON_OUTSIDE_UP] == 1 ||
			elevator.Requests[elevator.CurrentFloor][elevator_type.BUTTON_INSIDE_COMMAND] == 1 ||
			!requests_above(elevator) {
			return true
		}
	case elevator_type.Stop:
		return true
	}
	return false
}

func AddFloorToRequests(elevator elevator_type.ElevatorInfo, button elevator_type.ButtonInfo) elevator_type.ElevatorInfo {
	elevator.Requests[button.Floor][int(button.Button)] = 1
	return elevator
}

func ClearAtCurrentFloor(elevator elevator_type.ElevatorInfo) elevator_type.ElevatorInfo {
	for btn := 0; btn < elevator_type.N_BUTTONS; btn++ {
		elevator.Requests[elevator.CurrentFloor][btn] = 0
	}
	return elevator
}

func requests_above(elevator elevator_type.ElevatorInfo) bool {
	for floor := elevator.CurrentFloor + 1; floor < elevator_type.N_FLOORS; floor++ {
		for button := 0; button < elevator_type.N_BUTTONS; button++ {
			if elevator.Requests[floor][button] == 1 {
				return true
			}
		}
	}
	return false
}

func requests_below(elevator elevator_type.ElevatorInfo) bool {
	for floor := 0; floor < elevator.CurrentFloor; floor++ {
		for button := 0; button < elevator_type.N_BUTTONS; button++ {
			if elevator.Requests[floor][button] == 1 {
				return true
			}
		}
	}
	return false
}

func Requests_chooseDirection(elevator elevator_type.ElevatorInfo) elevator_type.Dir {
	switch elevator.Direction {
	case elevator_type.Up:
		if requests_above(elevator) {
			return elevator_type.Up
		} else if requests_below(elevator) {
			return elevator_type.Down
		} else {
			return elevator_type.Stop
		}

	case elevator_type.Down:
		if requests_below(elevator) {
			return elevator_type.Down
		} else if requests_above(elevator) {
			return elevator_type.Up
		} else {
			return elevator_type.Stop
		}

	case elevator_type.Stop:
		if requests_below(elevator) {
			return elevator_type.Down
		} else if requests_above(elevator) {
			return elevator_type.Up
		} else {
			return elevator_type.Stop
		}
	}

	return elevator_type.Up
}

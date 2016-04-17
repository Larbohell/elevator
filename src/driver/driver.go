package driver

import . "elevator_type"
import . "time"

// Make all driver funcs except Driver() lowercase

func elevator_init(doInit bool, startingPoint ElevatorInfo, floorSensorChannel chan int, errorChannel chan string, initialElevatorStateChannel chan ElevatorInfo) {
	Elevator_c_init()

	go read_floor_sensor(floorSensorChannel)

	elevator := startingPoint

	floor := <-floorSensorChannel

	if floor == -1 && elevator.State != State_Moving {
		Elevator_set_motor_direction(MOTOR_DIRECTION_DOWN)

	loop:
		for {
			select {

			case floor = <-floorSensorChannel:
				if floor != -1 {
					elevator.CurrentFloor = floor
					elevator.State = State_Idle
					Elevator_set_motor_direction(MOTOR_DIRECTION_STOP)
					break loop
				}

			case <-After(10 * Second):
				Elevator_set_motor_direction(MOTOR_DIRECTION_STOP)
				errorChannel <- "Elevator initialization failed. Timeout: Did not reach floor."
				return
			}
		}
	} else if elevator.State == State_Moving {
		Elevator_set_motor_direction(Motor_direction(elevator.Direction))
	}

	for floor = 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if floor != 0 && btn == int(BUTTON_OUTSIDE_DOWN) {
				Elevator_set_button_lamp(BUTTON_OUTSIDE_DOWN, floor, elevator.Requests[floor][btn])
			}
			if floor != (N_BUTTONS-1) && btn == int(BUTTON_OUTSIDE_UP) {
				Elevator_set_button_lamp(BUTTON_OUTSIDE_UP, floor, elevator.Requests[floor][btn])
			}
			Elevator_set_button_lamp(BUTTON_INSIDE_COMMAND, floor, elevator.Requests[floor][btn])
		}
	}
	initialElevatorStateChannel <- elevator
}

func Driver(doInit bool, startingPoint ElevatorInfo, setMovingDirectionChannel chan Dir, openDoorChannel chan bool, keepDoorOpenChannel chan bool, setButtonLightChannel chan ButtonInfo, newOrderChannel chan ButtonInfo, arrivedAtFloorChannel chan int, errorChannel chan string, initialElevatorStateChannel chan ElevatorInfo, doorClosedChannel chan bool, clearButtonLightsAtFloorChannel chan int) {
	floorSensorChannel := make(chan int, 1)
	elevator_init(doInit, startingPoint, floorSensorChannel, errorChannel, initialElevatorStateChannel)

	go lightsHandler(setButtonLightChannel, clearButtonLightsAtFloorChannel)
	go read_buttons(newOrderChannel)

	for {
		select {

		case movingDirection := <-setMovingDirectionChannel:
			Elevator_set_motor_direction(Motor_direction(movingDirection))

		case <-openDoorChannel:
			Elevator_set_motor_direction(MOTOR_DIRECTION_STOP)

			Elevator_set_door_open_lamp(1)

		loop:
			for {
				select {
				case <-keepDoorOpenChannel:
					break

				case <-After(3 * Second):
					break loop
				}
			}

			Elevator_set_door_open_lamp(0)
			doorClosedChannel <- true

		case floor := <-floorSensorChannel:
			if floor != -1 {
				Elevator_set_floor_indicator(floor)
				arrivedAtFloorChannel <- floor
			}
		}
	}
}

func lightsHandler(setButtonLightChannel chan ButtonInfo, clearButtonLightsAtFloorChannel chan int) {
	for {
		select {
		case buttonInfo := <-setButtonLightChannel:
			Elevator_set_button_lamp(buttonInfo.Button, buttonInfo.Floor, buttonInfo.Value)

		case floor := <-clearButtonLightsAtFloorChannel:
			for btn := 0; btn < N_BUTTONS; btn++ {
				Elevator_set_button_lamp(Button(btn), floor, 0)
			}
		}
	}
}

func read_buttons(newOrderChannel chan ButtonInfo) {
	var previous_button_value [N_FLOORS][N_BUTTONS]bool
	for {

		Sleep(80 * Millisecond)
		for floor := 0; floor < N_FLOORS; floor++ {
			for button := 0; button < N_BUTTONS; button++ {

				var button_value bool = Elevator_is_button_pushed(Button(button), floor)

				if button_value && button_value != previous_button_value[floor][button] {
					newOrder := ButtonInfo{Button(button), floor, 1}

					newOrderChannel <- newOrder
				}
				previous_button_value[floor][button] = button_value
			}
		}
	}
}

func read_floor_sensor(floorSensorChannel chan int) {
	lastFloor := N_FLOORS + 1 //Impossible floor value

	for {
		Sleep(10 * Millisecond)
		currentFloor := Elevator_get_floor_sensor_signal()
		if currentFloor != -1 {
		}
		if currentFloor != lastFloor {
			lastFloor = currentFloor
			floorSensorChannel <- currentFloor
		}
	}
}

func clearButtonLightsAtFloor(currentFloor int) {
	for floor := 0; floor < N_FLOORS; floor++ {
		if floor == currentFloor {
			for btn := 0; btn < N_BUTTONS; btn++ {
				Elevator_set_button_lamp(Button(btn), floor, 0)
			}
		}
	}
}

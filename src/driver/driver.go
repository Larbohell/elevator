package driver

import . "elevator_type"
import . "time"

import . "statusHandler"

//import "strconv"

// Make all driver funcs except Driver() lowercase

func elevator_init(doInit bool, floorSensorChannel chan int, errorChannel chan string, initialElevatorStateChannel chan ElevatorInfo, initIsFinished chan bool) {
	Elevator_c_init()

	if !doInit {
		initIsFinished <- true
		return
	}

	//Turn off all lights
	for floor := 0; floor < N_FLOORS; floor++ {
		if floor != 0 {
			Elevator_set_button_lamp(BUTTON_OUTSIDE_DOWN, floor, 0)
		}
		if floor != (N_BUTTONS - 1) {
			Elevator_set_button_lamp(BUTTON_OUTSIDE_UP, floor, 0)
		}
		Elevator_set_button_lamp(BUTTON_INSIDE_COMMAND, floor, 0)
	}
	Elevator_set_door_open_lamp(0)
	Elevator_set_floor_indicator(0)

	floor := <-floorSensorChannel

	if floor != -1 {
		//floorSensorChannel <- floor
		Elevator_set_motor_direction(MOTOR_DIRECTION_STOP)
	} else {
		Elevator_set_motor_direction(MOTOR_DIRECTION_DOWN)

	loop:
		for {
			//time.Sleep(10 * time.Millisecond)
			select {

			case floor = <-floorSensorChannel:
				//floorSensorChannel <- floor
				if floor != -1 {
					Elevator_set_motor_direction(MOTOR_DIRECTION_STOP)
					break loop
				}

			case <-After(10 * Second):
				Elevator_set_motor_direction(MOTOR_DIRECTION_STOP)
				errorChannel <- "Elevator initialization failed. Timeout: Did not reach floor."
				return
			}
		}
	}

	var initialRequests [N_FLOORS][N_BUTTONS]int
	elevator := ElevatorInfo{floor, Stop, initialRequests, State_Idle}
	initialElevatorStateChannel <- elevator

	initIsFinished <- true
}

func Driver(doInit bool, setMovingDirectionChannel chan Dir, openDoorChannel chan bool, setButtonLightChannel chan ButtonInfo, newOrderChannel chan ButtonInfo, initIsFinished chan bool, arrivedAtFloorChannel chan int, errorChannel chan string, initialElevatorStateChannel chan ElevatorInfo, doorClosedChannel chan bool, clearButtonLightsAtFloorChannel chan int) {
	floorSensorChannel := make(chan int, 1)
	go read_floor_sensor(floorSensorChannel)

	elevator_init(doInit, floorSensorChannel, errorChannel, initialElevatorStateChannel, initIsFinished)

	go read_buttons(newOrderChannel)

	for {
		select {

		case movingDirection := <-setMovingDirectionChannel:
			Elevator_set_motor_direction(Motor_direction(movingDirection))

		case <-openDoorChannel:
			StatusChannel <- "IN DRIVER, openDoorChannel"

			Elevator_set_motor_direction(MOTOR_DIRECTION_STOP)

			Elevator_set_door_open_lamp(1)

			<-After(3 * Second)
			Elevator_set_door_open_lamp(0)
			StatusChannel <- "IN DRIVER, openDoorChannel, door lamp should be off"
			doorClosedChannel <- true

		case buttonInfo := <-setButtonLightChannel:
			Elevator_set_button_lamp(buttonInfo.Button, buttonInfo.Floor, buttonInfo.Value)

		case floor := <-clearButtonLightsAtFloorChannel:
			for btn := 0; btn < N_BUTTONS; btn++ {
				Elevator_set_button_lamp(Button(btn), floor, 0)
			}

		case floor := <-floorSensorChannel:
			//StatusChannel <- "IN DRIVER: floorSensorChannel = " + strconv.Itoa(floor)
			if floor != -1 {
				Elevator_set_floor_indicator(floor)
				//StatusChannel <- "Before arrivedAtFloorChennel at floor " + strconv.Itoa(floor)
				arrivedAtFloorChannel <- floor
				//StatusChannel <- "Arrived at floor: " + strconv.Itoa(floor)

				//StatusChannel <- "Floor: " + strconv.Itoa(floor)
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
	//StatusChannel <- "Even here?"
	lastFloor := N_FLOORS + 1 //Impossible floor value

	for {
		Sleep(10 * Millisecond)
		//StatusChannel <- "Sleep"
		currentFloor := Elevator_get_floor_sensor_signal()
		if currentFloor != -1 {
			//StatusChannel <- "IN READ_FLOOR_SENSOR: Current floor = " + strconv.Itoa(currentFloor)
		}
		if currentFloor != lastFloor {
			lastFloor = currentFloor
			//StatusChannel <- "Before floorSensorChannel" + strconv.Itoa(currentFloor)
			floorSensorChannel <- currentFloor
			//StatusChannel <- "After floorSensorChannel, floor " + strconv.Itoa(currentFloor)
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

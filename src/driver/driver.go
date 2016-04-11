package driver
/*
#cgo CFLAGS: -std=c11
#cgo LDFLAGS: -lcomedi -lm
#include "elevator.h"
#include "io.h"
*/
import "C"
import "time"

type Motor_direction int

const (
	MOTOR_DIRECTION_DOWN Motor_direction = -1
	MOTOR_DIRECTION_UP Motor_direction = 1
	MOTOR_DIRECTION_STOP Motor_direction = 0
)

type Button int

const (
	BUTTON_OUTSIDE_UP Button = 0
	BUTTON_OUTSIDE_DOWN Button = 1
	BUTTON_INSIDE_COMMAND Button = 2
)

func Driver(setMovingDirectionChannel chan elevator_type.Dir, stopChannel chan bool, setButtonLightChannel chan elevator_type.ButtonInfo, newOrderChannel chan elevator_type.ButtonInfo, initIsFinished chan bool){
	floorSensorChannel := make(chan int, 1)
	go read_floor_sensor(sensorChannel)

	Elevator_init(sensorChannel)
	initIsFinished <- true
	
	go read_buttons(newOrderChannel)

	for {
		select {
		case movingDirection := <- setMovingDirectionChannel:
			Elevator_set_motor_direction(Motor_direction(movingDirection))

		case <- stopChannel:
			Elevator_set_motor_direction(MOTOR_DIRECTION_STOP)			
			Elevator_set_door_open_lamp(1);
			<- time.After(3*time.Second)
			Elevator_set_door_open_lamp(0);

		case buttonInfo := <- setButtonLightChannel:
			Elevator_set_button_lamp(buttonInfo.Button, buttonInfo.Floor, buttonInfo.Value)
		
		case floor := <- floorSensorChannel:
			if (floor != -1){
				Elevator_set_floor_indicator(floor)
				arrivedAtFloorChannel <- floor
			}
		}
	}
}

func read_buttons(newOrderChannel chan elevator_type.ButtonInfo){
	var previous_button_value[elevator_type.N_FLOORS][elevator_type.N_BUTTONS] bool
	for {

		time.Sleep(80*time.Millisecond)
		for floor := 0; floor < elevator_type.N_FLOORS; floor++{
			for button := 0; button < elevator_type.N_BUTTONS; button++{

				var button_value bool = Elevator_is_button_pushed(Button(button), floor)
				if (button_value && button_value != previous_button_value[floor][button]){
					newOrder elevator_type.ButtonInfo := {
						Button: Button(button),
						Floor: floor,
						Value: 1
					}
					newOrderChannel <- newOrder
				}
				previous_button_value[floor][button] = button_value

			}
		}
	}
}

func read_floor_sensor(floorSensorChannel chan int){
	lastFloor := -1 
	for {
		time.Sleep(10*time.Millisecond)
		currentFloor int := Elevator_get_floor_sensor_signal() 
		if (currentFloor != lastFloor)
			lastFloor = currentFloor
			floorSensorChannel <- currentFloor
	}
}

func Elevator_init(floorSensorChannel chan int, arrivedAtFloorChannel chan int){
	C.io_init()
	C.elev_init()

	//Turn off all lights
	for floor := 0; floor < elevator_type.N_FLOORS; floor++ {
		if floor != 0 {
			Elevator_set_button_lamp(BUTTON_OUTSIDE_DOWN, floor, 0)
		}
		if floor != (elevator_type.N_BUTTONS - 1) {
			Elevator_set_button_lamp(BUTTON_OUTSIDE_UP, floor, 0)
		}
		Elevator_set_button_lamp(BUTTON_INSIDE_COMMAND, floor, 0)
	}
	Elevator_set_door_open_lamp(0)
	Elevator_set_floor_indicator(0)

	floor int := <- floorSensorChannel
	floorSensorChannel <- floor
	if (floor == -1){
		Elevator_set_motor_direction(MOTOR_DIRECTION_DOWN)
		
		select{
			case floor := <- arrivedAtFloorChannel:
				arrivedAtFloorChannel <- floor
				return

			case <- time.After(10*time.Second):
				Elevator_set_motor_direction(MOTOR_DIRECTION_STOP)
				// Send error on error channel
				return
	}
}

func Elevator_set_motor_direction(direction Motor_direction){
	C.elev_set_motor_direction(C.elev_motor_direction_t(direction))
}

func Elevator_set_button_lamp(button Button, floor int, value int){
	C.elev_set_button_lamp(C.elev_button_type_t(button), C.int(floor), C.int(value))
}

func Elevator_set_floor_indicator(floor int){
	C.elev_set_floor_indicator(C.int(floor))
}

func Elevator_set_door_open_lamp(value int){
	C.elev_set_door_open_lamp(C.int(value))
}

func Elevator_set_stop_lamp(value int){
	C.elev_set_stop_lamp(C.int(value))
}

func Elevator_is_button_pushed(button Button, floor int) bool {
	return (C.elev_get_button_signal(C.elev_button_type_t(button), C.int(floor)) != 0)
}

func Elevator_get_floor_sensor_signal() int {
	return int(C.elev_get_floor_sensor_signal())
}

func Elevator_get_stop_signal() bool {
	return (int(C.elev_get_stop_signal()) != 0)
}

func Elevator_get_obstruction_signal() bool {
	return (int(C.elev_get_obstruction_signal()) != 0)
}
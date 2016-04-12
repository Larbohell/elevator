package driver

/*
#cgo CFLAGS: -std=c11
#cgo LDFLAGS: -lcomedi -lm
#include "elevator.h"
#include "io.h"
*/
import "C"
import "elevator_type"

//import "time"

type Motor_direction int

const (
	MOTOR_DIRECTION_DOWN Motor_direction = -1
	MOTOR_DIRECTION_UP   Motor_direction = 1
	MOTOR_DIRECTION_STOP Motor_direction = 0
)

func Elevator_c_init() {
	C.io_init()
	C.elev_init()

}

func Elevator_set_motor_direction(direction Motor_direction) {
	C.elev_set_motor_direction(C.elev_motor_direction_t(direction))
}

func Elevator_set_button_lamp(button elevator_type.Button, floor int, value int) {
	C.elev_set_button_lamp(C.elev_button_type_t(button), C.int(floor), C.int(value))
}

func Elevator_set_floor_indicator(floor int) {
	C.elev_set_floor_indicator(C.int(floor))
}

func Elevator_set_door_open_lamp(value int) {
	C.elev_set_door_open_lamp(C.int(value))
}

func Elevator_set_stop_lamp(value int) {
	C.elev_set_stop_lamp(C.int(value))
}

func Elevator_is_button_pushed(button elevator_type.Button, floor int) bool {
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

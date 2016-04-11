package elevator_type

const N_FLOORS int = 4
const N_BUTTONS int = 3

type Elevator_Behaviour int

const (
	EB_Idle Elevator_Behaviour = 0
	EB_DoorOpen Elevator_Behaviour = 1 
	EB_Moving Elevator_Behaviour = 2
)

type Dir int

const (
	Down Dir = -1
	Up Dir = 1
	Stop Dir = 0
)

type Elevator struct {
	Floor int
	Dir Dir
	Requests[N_FLOORS][N_BUTTONS] int
	Behaviour Elevator_Behaviour
}

type ButtonInfo struct {
	Button Button
	Floor int
	Value int
}

type Message struct {
	FromMaster bool
	ElevatorInfo Elevator
	ButtonInfo int
}
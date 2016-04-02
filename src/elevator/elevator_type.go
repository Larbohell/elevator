package elevator

const N_FLOORS int = 4
const N_BUTTONS int = 3

type Elevator_behaviour int

const (
	EB_Idle Elevator_behaviour = 0
	EB_DoorOpen Elevator_behaviour = 1 
	EB_Moving Elevator_behaviour = 2
)

type Dir int

const (
	Down Dir = -1
	Up Dir = 1
	Stop Dir = 0
)

type Elevator struct {
	floor int
	dir Dir
	requests[N_FLOORS][N_BUTTONS] int
	behaviour Elevator_behaviour
}
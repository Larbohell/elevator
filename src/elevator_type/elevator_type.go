package elevator_type

const N_FLOORS int = 4
const N_BUTTONS int = 3

type ElevatorState int

const (
	State_Idle     ElevatorState = 0
	State_DoorOpen ElevatorState = 1
	State_Moving   ElevatorState = 2
)

type Dir int

const (
	Down Dir = -1
	Up   Dir = 1
	Stop Dir = 0
)

type ElevatorInfo struct {
	CurrentFloor int
	Direction    Dir
	Requests     [N_FLOORS][N_BUTTONS]int
	State        ElevatorState
}

type Button int // Should be ButtonType

const (
	BUTTON_OUTSIDE_UP     Button = 0
	BUTTON_OUTSIDE_DOWN   Button = 1
	BUTTON_INSIDE_COMMAND Button = 2
)

type ButtonInfo struct {
	Button Button //should be named ButtonType Button
	Floor  int
	Value  int
}

type Message struct {
	FromMaster         bool
	NewOrder           bool
	ElevatorInfoUpdate bool

	ElevatorInfo ElevatorInfo
	ButtonInfo   ButtonInfo
}

package elevator

import "fmt"
import "elevator_type"

func Print_status(elevator elevator_type.ElevatorInfo) {
	fmt.Printf("****Elevator status**** \n")
	fmt.Printf("Floor: %d\n", elevator.CurrentFloor)

	switch elevator.Direction {
	case elevator_type.Up:
		fmt.Printf("Direction: Up\n")
	case elevator_type.Down:
		fmt.Printf("Direction: Down\n")
	case elevator_type.Stop:
		fmt.Printf("Direction: Stop\n")
	}

	switch elevator.State {
	case elevator_type.State_Idle:
		fmt.Printf("State: Idle\n")
	case elevator_type.State_Moving:
		fmt.Printf("State: Moving\n")
	case elevator_type.State_DoorOpen:
		fmt.Printf("State: Door Open\n")
	}

	fmt.Printf("\n")

	fmt.Printf("  +--------------------+\n")
	fmt.Printf("  |  | up  | dn  | cab |\n")
	for f := elevator_type.N_FLOORS - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := 0; btn < elevator_type.N_BUTTONS; btn++ {
			if f == elevator_type.N_FLOORS-1 && btn == int(elevator_type.BUTTON_OUTSIDE_UP) || f == 0 && btn == int(elevator_type.BUTTON_OUTSIDE_DOWN) {
				fmt.Printf("|     ")
			} else {
				if elevator.Requests[f][btn] == 1 {
					fmt.Printf("|  #  ")
				} else {
					fmt.Printf("|  -  ")
				}
			}
		}
		fmt.Printf("|\n")
	}
	fmt.Printf("  +--------------------+\n")
}

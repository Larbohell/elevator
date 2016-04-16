package elevator

import "fmt"
import . "elevator_type"

func Print_status(elevator ElevatorInfo) {
	fmt.Printf("\n****Elevator status**** \n")
	fmt.Printf("Floor: %d\n", elevator.CurrentFloor)

	switch elevator.Direction {
	case Up:
		fmt.Printf("Direction: Up\n")
	case Down:
		fmt.Printf("Direction: Down\n")
	case Stop:
		fmt.Printf("Direction: Stop\n")
	}
	/*
		switch(elevator.Behaviour){
		case elevator_type.EB_Idle:
			fmt.Printf("State: Idle\n")
		case elevator_type.EB_Moving:
			fmt.Printf("State: Moving\n")
		case elevator_type.EB_DoorOpen:
			fmt.Printf("State: Door Open\n")
		}
	*/
	fmt.Printf("\n")

	fmt.Printf("  +--------------------+\n")
	fmt.Printf("  |  | up  | dn  | cab |\n")
	for f := N_FLOORS - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := 0; btn < N_BUTTONS; btn++ {
			if f == N_FLOORS-1 && btn == int(BUTTON_OUTSIDE_UP) || f == 0 && btn == int(BUTTON_OUTSIDE_DOWN) {
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

func Print_external(externalOrders [N_FLOORS][N_BUTTONS - 1]string) {
	fmt.Printf("\n")

	fmt.Printf("  +--------------+\n")
	fmt.Printf("  |  | up  | dn  |\n")
	for f := N_FLOORS - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := 0; btn < N_BUTTONS-1; btn++ {
			if f == N_FLOORS-1 && btn == int(BUTTON_OUTSIDE_UP) || f == 0 && btn == int(BUTTON_OUTSIDE_DOWN) {
				fmt.Printf("|     ")
			}
		}
		fmt.Printf("|\n")
	}
	fmt.Printf("  +--------------+\n")
}

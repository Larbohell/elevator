package elevator

import "fmt"
import "driver"

func Print_status(elevator Elevator){
	fmt.Printf("****Elevator status**** \n")
	fmt.Printf("Floor: %d\n", elevator.floor )
	
	switch(elevator.dir){
	case Up:
		fmt.Printf("Direction: Up\n")
	case Down:
		fmt.Printf("Direction: Down\n")
	case Stop:
		fmt.Printf("Direction: Stop\n")
	}

	switch(elevator.behaviour){
	case EB_Idle:
		fmt.Printf("State: Idle\n")
	case EB_Moving:
		fmt.Printf("State: Moving\n")
	case EB_DoorOpen:
		fmt.Printf("State: Door Open\n")
	}
	
	fmt.Printf("\n")

	fmt.Printf("  +--------------------+\n");
    fmt.Printf("  |  | up  | dn  | cab |\n");
    for f := N_FLOORS-1; f >= 0; f--{
        fmt.Printf("  | %d", f);
        for btn := 0; btn < N_BUTTONS; btn++{
            if f == N_FLOORS-1 && btn == int(driver.BUTTON_OUTSIDE_UP) || f == 0 && btn == int(driver.BUTTON_OUTSIDE_DOWN) {
                fmt.Printf("|     ");
            } else {
            	if elevator.requests[f][btn] == 1{
            		fmt.Printf("|  #  ")
            	} else {
            		fmt.Printf("|  -  ")
            	}
            }
        }
        fmt.Printf("|\n");
    }
    fmt.Printf("  +--------------------+\n");
}
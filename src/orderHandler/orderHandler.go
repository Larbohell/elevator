package orderHandler

import . "elevator_type"
import . "statusHandler"

import "strconv"

//import "math"

func OrderHandler(addOrderChannel chan ButtonInfo, removeOrderChannel chan ButtonInfo, addToRequestsChannel chan ButtonInfo, externalOrderChannel chan ButtonInfo) {
	StatusChannel <- "In OrderHandler"

	for {
		select {
		case addOrder := <-addOrderChannel:
			if addOrder.Button == BUTTON_INSIDE_COMMAND {
				StatusChannel <- "	Button pushed = INTERNAL"
				addToRequestsChannel <- addOrder
			} else {
				// Inform master, master handles order and adds to one of the elevator's queues
				StatusChannel <- "	Button pushed = EXTERNAL"
				externalOrderChannel <- addOrder
			}
		}
	}
}

func BestElevatorForTheJob(findBestElevatorForTheJobChannel chan ButtonInfo, slavesAliveMapIsChangedChannel chan map[string]ElevatorInfo, thisIsTheBestElevatorChannel chan string, masterElevatorInfoChannel chan ElevatorInfo, masterIP string) string {
	elevatorsAliveMap := make(map[string]ElevatorInfo)

	var masterElevatorInfo ElevatorInfo

	for {
		select {
		case buttonInfo := <-findBestElevatorForTheJobChannel:
			//CAlculate
			minValue := N_FLOORS
			var bestElevatorIP string
			for elevatorIP, elevator := range elevatorsAliveMap {

				distance := elevator.CurrentFloor - buttonInfo.Floor
				// Jallaabs
				if distance < 0 {
					distance = distance * -1
				}
				if distance < minValue {
					minValue = distance
					bestElevatorIP = elevatorIP
				}
			}
			if minValue == N_FLOORS {
				StatusChannel <- "ERROR: No elevator is best for the job"
			}
			StatusChannel <- "Best IP for the job: " + bestElevatorIP
			StatusChannel <- "This elevator's distance to the requested floor is " + strconv.Itoa(minValue)
			/*
				StatusChannel <- strconv.Itoa(len(slavesAliveMap))
				StatusChannel <- "Button pushed in floor: " + strconv.Itoa(buttonInfo.Floor)
				elevatorIP := "129.241.187.159"
			*/
			thisIsTheBestElevatorChannel <- bestElevatorIP

		case elevatorsAliveMap = <-slavesAliveMapIsChangedChannel:
			elevatorsAliveMap[masterIP] = masterElevatorInfo

		case masterElevatorInfo = <-masterElevatorInfoChannel:
			elevatorsAliveMap[masterIP] = masterElevatorInfo

		}
	}
}

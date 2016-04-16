package orderHandler

import . "elevator_type"

//import . "statusHandler"

//import "strconv"

//import "math"

func OrderHandler(newOrderChannel chan ButtonInfo, removeOrderChannel chan ButtonInfo, addToRequestsChannel chan ButtonInfo, externalOrderChannel chan ButtonInfo) {
	//////StatusChannel <- "In OrderHandler"

	for {
		select {
		case addOrder := <-newOrderChannel:
			if addOrder.Button == BUTTON_INSIDE_COMMAND {
				//////StatusChannel <- "	Button pushed = INTERNAL, in floor: " + strconv.Itoa(addOrder.Floor)
				addToRequestsChannel <- addOrder
			} else {
				// Inform master, master handles order and adds to one of the elevator's queues
				//////StatusChannel <- "	Button pushed = EXTERNAL, in floor " + strconv.Itoa(addOrder.Floor)
				externalOrderChannel <- addOrder
				//////StatusChannel <- "Pushed to externalOrderCHannel"
			}
		}
	}
}

func BestElevatorForTheJob(findBestElevatorForTheJobChannel chan ButtonInfo, slavesAliveMapIsChangedChannel chan map[string]ElevatorInfo, thisIsTheBestElevatorChannel chan string, masterElevatorInfoChannel chan ElevatorInfo, masterIP string, terminateThreadChannel chan bool, threadIsTerminatedChannel chan bool) {
	slavesAliveMap := make(map[string]ElevatorInfo)

	var masterElevatorInfo ElevatorInfo

	for {
		select {
		case buttonInfo := <-findBestElevatorForTheJobChannel:

			minValue := N_FLOORS
			var bestElevatorIP string

			for slaveIP, elevator := range slavesAliveMap {
				slaveCostValue := costFunction(elevator, buttonInfo)

				if slaveCostValue < minValue {
					minValue = slaveCostValue
					bestElevatorIP = slaveIP
				}
			}
			//////StatusChannel <- "Slave distance: " + strconv.Itoa(minValue)

			masterCostValue := costFunction(masterElevatorInfo, buttonInfo)
			//////StatusChannel <- "Master distance: " + strconv.Itoa(masterCostValue)
			if masterCostValue < minValue {
				minValue = masterCostValue
				bestElevatorIP = masterIP
			}

			if minValue == N_FLOORS {
				//////StatusChannel <- "ERROR: No elevator is best for the job"
			}

			//////StatusChannel <- "Best IP for the job: " + bestElevatorIP
			//////StatusChannel <- "This elevator's distance to the requested floor is " + strconv.Itoa(minValue)
			/*
				StatusChannel <- strconv.Itoa(len(slavesAliveMap))
				StatusChannel <- "Button pushed in floor: " + strconv.Itoa(buttonInfo.Floor)
				elevatorIP := "129.241.187.159"
			*/
			thisIsTheBestElevatorChannel <- bestElevatorIP

		case slavesAliveMap = <-slavesAliveMapIsChangedChannel:
			break //update elevatorsAliveMap

		case masterElevatorInfo = <-masterElevatorInfoChannel:
			break

		case terminate := <-terminateThreadChannel:
			terminateThreadChannel <- terminate
			threadIsTerminatedChannel <- true
			//////StatusChannel <- "BestElevatorForTheJob is terminated"
			return
			//elevatorsAliveMap[masterIP] = masterElevatorInfo

		}
	}
}

func costFunction(elevator ElevatorInfo, buttonInfo ButtonInfo) int {
	var cost int
	var directionToOrder Dir
	distance := elevator.CurrentFloor - buttonInfo.Floor
	// JallaAbs()
	if distance < 0 {
		distance = distance * -1
		directionToOrder = Down
	} else {
		directionToOrder = Up
	}

	if elevator.Direction != Stop {
		if directionToOrder != elevator.Direction {
			cost += N_FLOORS
		}
	}
	cost += distance
	return cost
}

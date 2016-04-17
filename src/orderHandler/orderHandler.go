package orderHandler

import . "elevator_type"
import . "statusHandler"

import . "strconv"

//import "math"

func OrderHandler(addOrderChannel chan ButtonInfo, removeOrderChannel chan ButtonInfo, addToRequestsChannel chan ButtonInfo, externalOrderChannel chan ButtonInfo) {
	counter := 0
	StatusChannel <- Itoa(counter) + ":In OrderHandler"

	for {
		select {
		case addOrder := <-addOrderChannel:

			if addOrder.Button == BUTTON_INSIDE_COMMAND {

				StatusChannel <- "	Button pushed = INTERNAL"
				addToRequestsChannel <- addOrder
			} else {
				counter++
				// Inform master, master handles order and adds to one of the elevator's queues
				StatusChannel <- Itoa(counter) + ": Button pushed = EXTERNAL"
				externalOrderChannel <- addOrder
				StatusChannel <- Itoa(counter) + ": Button pushed = EXTERNAL and put to EXTERNALORDERCHANNEL"
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

			masterCostValue := costFunction(masterElevatorInfo, buttonInfo)
			if masterCostValue < minValue {
				minValue = masterCostValue
				bestElevatorIP = masterIP
			}

			if minValue == N_FLOORS {
				StatusChannel <- "ERROR: No elevator is best for the job"
			}

			//StatusChannel <- "Best IP for the job: " + bestElevatorIP
			//StatusChannel <- "This elevator's distance to the requested floor is " + strconv.Itoa(minValue)
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
			StatusChannel <- "BestElevatorForTheJob is terminated"
			return
			//elevatorsAliveMap[masterIP] = masterElevatorInfo

		}
	}
}

func costFunction(elevator ElevatorInfo, buttonInfo ButtonInfo) int {
	distance := elevator.CurrentFloor - buttonInfo.Floor
	// JallaAbs()
	if distance < 0 {
		distance = distance * -1
	}
	return distance
}

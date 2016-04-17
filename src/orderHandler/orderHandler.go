package orderHandler

import . "elevator_type"
import . "statusHandler"

import . "strconv"

func OrderHandler(addOrderChannel chan ButtonInfo, removeOrderChannel chan ButtonInfo, addToRequestsChannel chan ButtonInfo, externalOrderChannel chan ButtonInfo) {
	counter := 0
	for {
		select {
		case addOrder := <-addOrderChannel:

			if addOrder.Button == BUTTON_INSIDE_COMMAND {

				StatusChannel <- "	Button pushed = INTERNAL"
				addToRequestsChannel <- addOrder
			} else {
				counter++
				StatusChannel <- Itoa(counter) + ": Button pushed = EXTERNAL"
				externalOrderChannel <- addOrder
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

			minValue := N_FLOORS * 2 // Max possible value of cost function + 1
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

			thisIsTheBestElevatorChannel <- bestElevatorIP

		case slavesAliveMap = <-slavesAliveMapIsChangedChannel:
			break

		case masterElevatorInfo = <-masterElevatorInfoChannel:
			break

		case terminate := <-terminateThreadChannel:
			terminateThreadChannel <- terminate
			threadIsTerminatedChannel <- true
			StatusChannel <- "BestElevatorForTheJob is terminated"
			return

		}
	}
}

func costFunction(elevator ElevatorInfo, buttonInfo ButtonInfo) int {
	var cost int

	var directionToOrder Dir
	distance := elevator.CurrentFloor - buttonInfo.Floor

	if distance < 0 {
		distance = distance * -1
		directionToOrder = Up
	} else {
		directionToOrder = Down
	}

	if elevator.Direction != Stop {
		if directionToOrder != elevator.Direction {
			cost += N_FLOORS
		}
	}
	cost += distance
	StatusChannel <- "Distance = " + Itoa(distance) + " and directionToOrder = " + Itoa(int(directionToOrder)) + " and elevator.Direction = " + Itoa(int(elevator.Direction))
	StatusChannel <- "Cost = " + Itoa(cost)
	return cost
}

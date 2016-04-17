package elevator

import "driver"
import . "elevator_type"
import . "statusHandler"
import "orderHandler"

import "network"
import "fileHandler"

import "strconv"

func Run_elevator(startingPoint ElevatorInfo, errorChannel chan string, programAliveChannel chan bool) {

	var elevator ElevatorInfo
	var uncompletedExternalOrders [N_FLOORS][N_BUTTONS - 1]string

	previousFloor := N_FLOORS + 1

	//Driver
	setMovingDirectionChannel := make(chan Dir, 1)
	openDoorChannel := make(chan bool, 1)
	keepDoorOpenChannel := make(chan bool, 1)
	setButtonLightChannel := make(chan ButtonInfo, 1)
	clearButtonLightsAtFloorChannel := make(chan int, 1)
	initialElevatorStateChannel := make(chan ElevatorInfo)

	//Orders
	newOrderChannel := make(chan ButtonInfo, 1)
	removeOrderChannel := make(chan ButtonInfo, 1)

	//States
	addToRequestsChannel := make(chan ButtonInfo, 1)
	stop := make(chan bool, 1)
	doorClosedChannel := make(chan bool, 1)
	arrivedAtFloorChannel := make(chan int, 1)
	//uncompletedExternalOrdersMatrixChangedChannel := make(chan [N_FLOORS][N_BUTTONS - 1]string, 1)
	updateExternalLightsChannel := make(chan [N_FLOORS][N_BUTTONS - 1]string, 1)

	//network channels
	orderCompletedByThisElevatorChannel := make(chan ButtonInfo, 1)
	externalOrderChannel := make(chan ButtonInfo, N_FLOORS*2-2) //N_FLOORS*2-2 = number of external buttons

	updateElevatorInfoChannel := make(chan ElevatorInfo, 1)

	// File handling
	backupChannel := make(chan ElevatorInfo, 1)

	elevator = startingPoint

	go driver.Driver(elevator, setMovingDirectionChannel, openDoorChannel, keepDoorOpenChannel, setButtonLightChannel, newOrderChannel, arrivedAtFloorChannel, errorChannel, initialElevatorStateChannel, doorClosedChannel, clearButtonLightsAtFloorChannel)

	elevator = <-initialElevatorStateChannel

	updateElevatorInfoChannel <- elevator
	backupChannel <- elevator

	go network.Slave(elevator, externalOrderChannel, updateElevatorInfoChannel, addToRequestsChannel, uncompletedExternalOrders, orderCompletedByThisElevatorChannel, updateExternalLightsChannel)
	go orderHandler.OrderHandler(newOrderChannel, removeOrderChannel, addToRequestsChannel, externalOrderChannel)
	go fileHandler.BackupElevatorInfoToFile(backupChannel)

	switch elevator.State {
	case State_DoorOpen:
		stop <- true
		break

	case State_Idle:
		break
	case State_Moving:
		break
	}
	counter := 0
	for {
		select {
		case buttonPushed := <-addToRequestsChannel:
			counter++
			StatusChannel <- strconv.Itoa(counter) + ": addToRequestsChannel"
			if buttonPushed.Button == BUTTON_INSIDE_COMMAND {
				setButtonLightChannel <- buttonPushed
			}

			switch elevator.State {

			case State_Idle:

				if elevator.CurrentFloor != buttonPushed.Floor {

					elevator = orderHandler.AddFloorToRequests(elevator, buttonPushed)
					direction := buttonPushed.Floor - elevator.CurrentFloor
					if direction > 0 {
						direction = int(Up)
					} else if direction < 0 {
						direction = int(Down)
					} else {
						errorChannel <- "In buttonPushed-> State_Idle: Direction is zero"
					}
					setMovingDirectionChannel <- Dir(direction)
					elevator.Direction = Dir(direction)
					elevator.State = State_Moving

					updateElevatorInfoChannel <- elevator
				} else {
					if buttonPushed.Button != BUTTON_INSIDE_COMMAND {
						orderCompletedByThisElevatorChannel <- buttonPushed
					}
					stop <- true
				}
				backupChannel <- elevator

			case State_Moving:
				elevator = orderHandler.AddFloorToRequests(elevator, buttonPushed)

				updateElevatorInfoChannel <- elevator
				backupChannel <- elevator

			case State_DoorOpen:

				if elevator.CurrentFloor == buttonPushed.Floor {
					keepDoorOpenChannel <- true
					if buttonPushed.Button != BUTTON_INSIDE_COMMAND {
						orderCompletedByThisElevatorChannel <- buttonPushed
					}
				} else {
					elevator = orderHandler.AddFloorToRequests(elevator, buttonPushed)
				}

				clearButtonLightsAtFloorChannel <- elevator.CurrentFloor

				updateElevatorInfoChannel <- elevator
				backupChannel <- elevator
			}
			programAliveChannel <- true

		case <-stop:
			StatusChannel <- "	stop"
			if elevator.State == State_Moving {
				if elevator.Requests[elevator.CurrentFloor][int(BUTTON_OUTSIDE_UP)] == 1 {
					var button ButtonInfo
					button.Button = BUTTON_OUTSIDE_UP
					button.Floor = elevator.CurrentFloor
					button.Value = 1
					orderCompletedByThisElevatorChannel <- button
				}
				if elevator.Requests[elevator.CurrentFloor][int(BUTTON_OUTSIDE_DOWN)] == 1 {
					var button ButtonInfo
					button.Button = BUTTON_OUTSIDE_DOWN
					button.Floor = elevator.CurrentFloor
					button.Value = 1

					orderCompletedByThisElevatorChannel <- button
				}
			}

			clearButtonLightsAtFloorChannel <- elevator.CurrentFloor
			openDoorChannel <- true

			elevator.State = State_DoorOpen

			elevator = orderHandler.ClearAtCurrentFloor(elevator)

			updateElevatorInfoChannel <- elevator
			backupChannel <- elevator
			programAliveChannel <- true

		case <-doorClosedChannel:
			StatusChannel <- "	doorClosedChannel"
			elevator.Direction = orderHandler.Requests_chooseDirection(elevator)
			setMovingDirectionChannel <- elevator.Direction

			if elevator.Direction != Stop {
				elevator.State = State_Moving
			} else {
				elevator.State = State_Idle
			}
			updateElevatorInfoChannel <- elevator
			backupChannel <- elevator
			programAliveChannel <- true

		case arrivedAtFloor := <-arrivedAtFloorChannel:
			StatusChannel <- "	arrivedAtFloorChannel"

			previousFloor = elevator.CurrentFloor
			elevator.CurrentFloor = arrivedAtFloor

			direction := elevator.CurrentFloor - previousFloor

			if direction < 0 {
				elevator.Direction = Down
			} else if direction > 0 {
				elevator.Direction = Up
			} else if direction == 0 {
				elevator.Direction = Stop
			} else {
				errorChannel <- "Error in case 'arriwedAtFloor': Direction neither up nor down."
			}

			if orderHandler.ShouldStop(elevator) {
				stop <- true
			}

			updateElevatorInfoChannel <- elevator
			backupChannel <- elevator
			programAliveChannel <- true

		case uncompletedExternalOrders := <-updateExternalLightsChannel:
			StatusChannel <- "uncompletedExternalOrderMatrixChangedChannel"
			for floor := 0; floor < N_FLOORS; floor++ {
				for btn := 0; btn < N_BUTTONS-1; btn++ {
					var button ButtonInfo
					button.Button = Button(btn)
					button.Floor = floor

					if uncompletedExternalOrders[floor][btn] == "" {
						button.Value = 0
					} else {
						button.Value = 1
					}

					setButtonLightChannel <- button
					StatusChannel <- "Updated light, floor: " + strconv.Itoa(floor) + "; button: " + strconv.Itoa(btn) + "; Value: " + strconv.Itoa(button.Value)
				}
			}
			programAliveChannel <- true
		}
	}
}

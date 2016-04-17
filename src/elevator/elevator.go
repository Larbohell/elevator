package elevator

import "driver"
import . "elevator_type"
import . "statusHandler"
import "orderHandler"

import "network"
import "fileHandler"

//import "net"

//import "fmt"

import "strconv"

//"129.241.187.156" = workspace 9
//"129.24.187.159" = workspace 11
//"129.24.187.152" = workspace 13
//"129.24.187.158" = workspace 10

//TODO!!!

// Light syncing is buggy
// On init: Check floor signal and compare with CUrrentFloor from file (elevator sometimes thinks it's somewhere else after init)

func Run_elevator(firstTimeRunning bool, startingPoint ElevatorInfo, errorChannel chan string) {

	var elevator ElevatorInfo
	var uncompletedExternalOrders [N_FLOORS][N_BUTTONS - 1]string //Could be declared in Slave

	previousFloor := N_FLOORS + 1 // Impossible floor

	setMovingDirectionChannel := make(chan Dir, 1)
	openDoorChannel := make(chan bool, 1)
	keepDoorOpenChannel := make(chan bool, 1)
	setButtonLightChannel := make(chan ButtonInfo, 1)
	clearButtonLightsAtFloorChannel := make(chan int, 1)

	//Debugging

	newOrderChannel := make(chan ButtonInfo, 1)
	removeOrderChannel := make(chan ButtonInfo, 1)
	arrivedAtFloorChannel := make(chan int, 1)

	addToRequestsChannel := make(chan ButtonInfo, 1)
	stop := make(chan bool, 1)
	initialElevatorStateChannel := make(chan ElevatorInfo)
	doorClosedChannel := make(chan bool, 1)

	//network channels
	orderCompletedByThisElevatorChannel := make(chan ButtonInfo, 1)

	//___________________________________________////////////////////////////////////////////////////////_____________________________
	externalOrderChannel := make(chan ButtonInfo, N_FLOORS*2-2) //N_FLOORS*2-2 = number of external buttons
	///////////////////////////////////////____________--------------------************************************************************

	updateElevatorInfoChannel := make(chan ElevatorInfo, 1)
	uncompletedExternalOrdersMatrixChangedChannel := make(chan [N_FLOORS][N_BUTTONS - 1]string, 1)

	// File handling
	backupChannel := make(chan ElevatorInfo, 1)

	elevator = startingPoint

	if firstTimeRunning {
		go driver.Driver(false, elevator, setMovingDirectionChannel, openDoorChannel, keepDoorOpenChannel, setButtonLightChannel, newOrderChannel, arrivedAtFloorChannel, errorChannel, initialElevatorStateChannel, doorClosedChannel, clearButtonLightsAtFloorChannel)

		//elevator = <-initialElevatorStateChannel
	} else {

		// Run driver with startingPoint
		go driver.Driver(false, elevator, setMovingDirectionChannel, openDoorChannel, keepDoorOpenChannel, setButtonLightChannel, newOrderChannel, arrivedAtFloorChannel, errorChannel, initialElevatorStateChannel, doorClosedChannel, clearButtonLightsAtFloorChannel)
	}

	elevator = <-initialElevatorStateChannel

	updateElevatorInfoChannel <- elevator
	backupChannel <- elevator
	//Running threads

	go network.Slave(elevator, externalOrderChannel, updateElevatorInfoChannel, addToRequestsChannel, uncompletedExternalOrders, orderCompletedByThisElevatorChannel, uncompletedExternalOrdersMatrixChangedChannel)
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
		//StatusChannel <- "In main select: "
		select {
		case buttonPushed := <-addToRequestsChannel:
			counter++
			StatusChannel <- strconv.Itoa(counter) + ": addToRequestsChannel"
			if buttonPushed.Button == BUTTON_INSIDE_COMMAND {
				setButtonLightChannel <- buttonPushed
			}

			switch elevator.State {

			case State_Idle:
				StatusChannel <- "Case State_Idle"
				//StatusChannel <- strconv.Itoa(counter) + ": State: Idle\n"
				//elevator = orderHandler.AddFloorToRequests(elevator, buttonPushed)
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
					//StatusChannel <- strconv.Itoa(counter) + ": Elevator Idle in same floor as button pushed"

					/* !!!!!!!!!!!!!!!!!!!!!!!!!!!
					if buttonPushed.Button != BUTTON_INSIDE_COMMAND {
						orderCompletedByThisElevatorChannel <- buttonPushed
					}
					!!!!!!!!!!!!!!!!!!! */
					stop <- true
				}
				backupChannel <- elevator

			case State_Moving:
				StatusChannel <- "Case State Moving"
				//StatusChannel <- strconv.Itoa(counter) + ": State: Moving\n"
				elevator = orderHandler.AddFloorToRequests(elevator, buttonPushed)

				updateElevatorInfoChannel <- elevator
				backupChannel <- elevator

			case State_DoorOpen:
				StatusChannel <- "Case State_DoorOpen"
				//TODO: If button pushed in same floor, do stop. When button pushed and Door open, button light doesn't turn on until door is closed
				//StatusChannel <- strconv.Itoa(counter) + ": State: DoorOpen\n"
				//elevator = orderHandler.AddFloorToRequests(elevator, buttonPushed)
				/*
					if elevator.CurrentFloor != buttonPushed.Floor {
						elevator = orderHandler.AddFloorToRequests(elevator, buttonPushed)
					} else {
						stop <- false // Should not open door
					}
				*/
				if elevator.CurrentFloor == buttonPushed.Floor {
					keepDoorOpenChannel <- true
					if buttonPushed.Button != BUTTON_INSIDE_COMMAND {
						orderCompletedByThisElevatorChannel <- buttonPushed
					}
				} else {
					elevator = orderHandler.AddFloorToRequests(elevator, buttonPushed)
				}
				/*
					if elevator.CurrentFloor == buttonPushed.Floor && buttonPushed.Button != BUTTON_INSIDE_COMMAND {
						orderCompletedByThisElevatorChannel <- buttonPushed
						//stop <- false // Should not open door
					} else {
						elevator = orderHandler.AddFloorToRequests(elevator, buttonPushed)
					}
				*/
				clearButtonLightsAtFloorChannel <- elevator.CurrentFloor

				updateElevatorInfoChannel <- elevator
				backupChannel <- elevator
			}

		case <-stop:
			StatusChannel <- "	stop"
			//if elevator.State == State_Moving {
			if elevator.State != State_DoorOpen {
				if elevator.Requests[elevator.CurrentFloor][int(BUTTON_OUTSIDE_UP)] == 1 {
					var button ButtonInfo
					button.Button = BUTTON_OUTSIDE_UP
					button.Floor = elevator.CurrentFloor
					button.Value = 1
					orderCompletedByThisElevatorChannel <- button
					StatusChannel <- "In Stop, orderCompletedByThisElevatorChannel BUTTON OUTSIDE UP"
				}
				if elevator.Requests[elevator.CurrentFloor][int(BUTTON_OUTSIDE_DOWN)] == 1 {
					var button ButtonInfo
					button.Button = BUTTON_OUTSIDE_DOWN
					button.Floor = elevator.CurrentFloor
					button.Value = 1

					orderCompletedByThisElevatorChannel <- button
					StatusChannel <- "In Stop, orderCompletedByThisElevatorChannel BUTTON OUTSIDE DOWN"

				}
			}

			clearButtonLightsAtFloorChannel <- elevator.CurrentFloor
			openDoorChannel <- true
			//elevator.State = State_Idle
			//StatusChannel <- "after openDoorChannel done"

			elevator.State = State_DoorOpen

			elevator = orderHandler.ClearAtCurrentFloor(elevator)
			//StatusChannel <- "After ClearAtCurrentFloor"

			updateElevatorInfoChannel <- elevator
			backupChannel <- elevator
			StatusChannel <- "stop case DONE"

		case <-doorClosedChannel:
			StatusChannel <- "	doorClosedChannel"
			elevator.Direction = orderHandler.Requests_chooseDirection(elevator)
			setMovingDirectionChannel <- elevator.Direction

			if elevator.Direction != Stop {
				elevator.State = State_Moving
			} else {
				elevator.State = State_Idle
			}
			StatusChannel <- "1"
			updateElevatorInfoChannel <- elevator
			StatusChannel <- "2"
			backupChannel <- elevator
			StatusChannel <- "3"

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
				elevator.Direction = Stop // Happens first time, after init
			} else {
				errorChannel <- "Error in case 'arriwedAtFloor': Direction neither up nor down."
			}

			if orderHandler.ShouldStop(elevator) {
				stop <- true
			}

			updateElevatorInfoChannel <- elevator
			backupChannel <- elevator

		case uncompletedExternalOrders := <-uncompletedExternalOrdersMatrixChangedChannel: //change to updateExtLightsChannel
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
					StatusChannel <- "Updated light, floor: " + strconv.Itoa(floor) + "; button: " + strconv.Itoa(btn) + "; to value: " + strconv.Itoa(button.Value)
				}
			}

		}
	}
}

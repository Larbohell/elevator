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
//Open door when elevator idle and buttons pushed in same floor, and turn off lights
// Aggressive button push on external buttons in floor 0 or 3 makes the elevator pass that floor and go out of bounds
// Process-pairs thing, write info to file so that backup can take over where the main process died
//uncompletedExternalList is not received by Slaves!

// The problems above might have been solved by updating elevator in Master thread, check this

// In driver.read_floor_sensor() -> After backup process takes over, floor sensor never returns a value != -1, which means that arrivedAtFloorCHannel in elevator.go never activates
// so the elevator never stops at a floor

// Button light don't light up before after door close

//Case DoorOpen: lights

func Run_elevator(firstTimeRunning bool, startingPoint ElevatorInfo, errorChannel chan string) {

	//const localIP string = "129.241.187.156" //workspace 11
	var elevator ElevatorInfo
	var uncompletedExternalOrders [N_FLOORS][N_BUTTONS - 1]string //Could be declared in Slave

	previousFloor := N_FLOORS + 1 // Impossible floor

	setMovingDirectionChannel := make(chan Dir, 1)
	openDoorChannel := make(chan bool, 1)
	setButtonLightChannel := make(chan ButtonInfo, 1)
	clearButtonLightsAtFloorChannel := make(chan int, 1)

	//Debugging
	//errorChannel := make(chan string)
	//StatusChannel = make(chan string)

	newOrderChannel := make(chan ButtonInfo, 1)
	removeOrderChannel := make(chan ButtonInfo, 1)
	initIsFinished := make(chan bool)
	initFloorChannel := make(chan int)
	arrivedAtFloorChannel := make(chan int, 1)

	addToRequestsChannel := make(chan ButtonInfo)
	stop := make(chan bool, 1)
	//initialElevatorStateChannel := make(chan ElevatorInfo, 1)
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

	//init
	//go Error_handler(errorChannel)
	//go Status_handler()

	elevator = startingPoint

	if firstTimeRunning {
		StatusChannel <- "First time running, starting normal init"
		go driver.Driver(false, elevator, setMovingDirectionChannel, openDoorChannel, setButtonLightChannel, newOrderChannel, initIsFinished, arrivedAtFloorChannel, errorChannel, initFloorChannel, doorClosedChannel, clearButtonLightsAtFloorChannel)

		//elevator = <-initialElevatorStateChannel
		StatusChannel <- "Current floor on first init = " + strconv.Itoa(elevator.CurrentFloor)
	} else {
		StatusChannel <- "Not first time running, beginning recovery from last session"
		StatusChannel <- "Current floor on init from backup process = " + strconv.Itoa(elevator.CurrentFloor)

		// Run driver with startingPoint
		go driver.Driver(false, elevator, setMovingDirectionChannel, openDoorChannel, setButtonLightChannel, newOrderChannel, initIsFinished, arrivedAtFloorChannel, errorChannel, initFloorChannel, doorClosedChannel, clearButtonLightsAtFloorChannel)
	}
	//elevator = <-initialElevatorStateChannel
	//<-initIsFinished
	elevator.CurrentFloor = <-initFloorChannel

	StatusChannel <- "Recovery/init done"
	StatusChannel <- "IN ELEVATOR: Elevator currentFloor = " + strconv.Itoa(elevator.CurrentFloor) + " , direction = " + strconv.Itoa(int(elevator.Direction)) + ", State: " + strconv.Itoa(int(elevator.State))

	updateElevatorInfoChannel <- elevator
	backupChannel <- elevator
	//Running threads

	go network.Slave(elevator, externalOrderChannel, updateElevatorInfoChannel, addToRequestsChannel, uncompletedExternalOrders, orderCompletedByThisElevatorChannel, uncompletedExternalOrdersMatrixChangedChannel)
	//go network.Master(elevator, externalOrderChannel, updateElevatorInfoChannel, addToRequestsChannel)
	go orderHandler.OrderHandler(newOrderChannel, removeOrderChannel, addToRequestsChannel, externalOrderChannel)
	go fileHandler.BackupElevatorInfoToFile(backupChannel)

	//if !firstTimeRunning {
	switch elevator.State {
	case State_DoorOpen:
		StatusChannel <- "On init not first time running, Door Open"
		stop <- true
		break

	case State_Idle:
		break
	case State_Moving:
		break
	}
	//}

	for {
		StatusChannel <- "In main select: "
		select {
		case buttonPushed := <-addToRequestsChannel:

			StatusChannel <- "	addToRequestsChannel, "
			setButtonLightChannel <- buttonPushed
			switch elevator.State {

			case State_Idle:
				StatusChannel <- "		State: Idle\n"
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
					StatusChannel <- "			Elevator Idle in same floor as button pushed"
					//stop <- true
				}
				backupChannel <- elevator

			case State_Moving:
				StatusChannel <- "		State: Moving\n"
				elevator = orderHandler.AddFloorToRequests(elevator, buttonPushed)

				updateElevatorInfoChannel <- elevator
				backupChannel <- elevator

			case State_DoorOpen:
				//TODO: If button pushed in same floor, do stop. When button pushed and Door open, button light doesn't turn on until door is closed
				StatusChannel <- "		State: DoorOpen\n"
				elevator = orderHandler.AddFloorToRequests(elevator, buttonPushed)

				updateElevatorInfoChannel <- elevator
				backupChannel <- elevator

			}

		case <-stop:
			StatusChannel <- "	stop"

			openDoorChannel <- true
			//elevator.State = State_Idle
			StatusChannel <- "after openDoorChannel done"
			elevator.State = State_DoorOpen

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

			elevator = orderHandler.ClearAtCurrentFloor(elevator)
			StatusChannel <- "After ClearAtCurrentFloor"

			clearButtonLightsAtFloorChannel <- elevator.CurrentFloor
			StatusChannel <- "1"
			updateElevatorInfoChannel <- elevator
			StatusChannel <- "2"
			backupChannel <- elevator
			StatusChannel <- "3"

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

		case arrivedAtFloor := <-arrivedAtFloorChannel:
			StatusChannel <- "	arrivedAtFloorChannel"

			previousFloor = elevator.CurrentFloor
			elevator.CurrentFloor = arrivedAtFloor

			//handleOrdersChannel <- elevator.CurrentFloor
			//nextOrder := <-getNextOrderChannel

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
				}
			}
		}
	}
}

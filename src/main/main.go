package main

import "driver"
import . "elevator_type"
import . "statusHandler"
import "orderHandler"

import "network"

//import "net"

//import "fmt"

//import . "strconv"

//"129.241.187.156" = workspace 9
//"129.24.187.159" = workspace 11
//"129.24.187.152" = workspace 13
//"129.24.187.158" = workspace 10

func main() {

	//const localIP string = "129.241.187.156" //workspace 11

	var elevator ElevatorInfo
	previousFloor := N_FLOORS + 1 // Impossible floor

	setMovingDirectionChannel := make(chan Dir, 1)
	stopChannel := make(chan bool, 1)
	setButtonLightChannel := make(chan ButtonInfo, 1)
	clearButtonLightsAtFloorChannel := make(chan int, 1)

	//Debugging
	errorChannel := make(chan string)
	StatusChannel = make(chan string)

	newOrderChannel := make(chan ButtonInfo, 1)
	removeOrderChannel := make(chan ButtonInfo, 1)
	initIsFinished := make(chan bool)
	arrivedAtFloorChannel := make(chan int, 1)

	addToRequestsChannel := make(chan ButtonInfo)
	stop := make(chan bool, 1)
	initialElevatorStateChannel := make(chan ElevatorInfo, 1)
	doorClosedChannel := make(chan bool, 1)

	//network channels
	externalOrderChannel := make(chan ButtonInfo, 1)
	updateElevatorInfoChannel := make(chan ElevatorInfo, 1)

	//init
	go Error_handler(errorChannel)
	go Status_handler()
	go driver.Driver(setMovingDirectionChannel, stopChannel, setButtonLightChannel, newOrderChannel, initIsFinished, arrivedAtFloorChannel, errorChannel, initialElevatorStateChannel, doorClosedChannel, clearButtonLightsAtFloorChannel)
	<-initIsFinished
	elevator = <-initialElevatorStateChannel
	updateElevatorInfoChannel <- elevator

	//Running threads
	go network.Slave(elevator, externalOrderChannel, updateElevatorInfoChannel, addToRequestsChannel)
	//go network.Master(elevator, externalOrderChannel, updateElevatorInfoChannel, addToRequestsChannel)
	go orderHandler.OrderHandler(newOrderChannel, removeOrderChannel, addToRequestsChannel, externalOrderChannel)

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
					elevator.State = State_Moving

					updateElevatorInfoChannel <- elevator
				} else {
					StatusChannel <- "			Elevator Idle in same floor as button pushed"
					//stop <- true
				}

			case State_Moving:
				StatusChannel <- "		State: Moving\n"
				elevator = orderHandler.AddFloorToRequests(elevator, buttonPushed)

				updateElevatorInfoChannel <- elevator
			}

		case <-stop:
			StatusChannel <- "	stop"

			stopChannel <- true
			elevator.State = State_Idle
			elevator = orderHandler.ClearAtCurrentFloor(elevator)
			clearButtonLightsAtFloorChannel <- elevator.CurrentFloor

			updateElevatorInfoChannel <- elevator

		case <-doorClosedChannel:
			StatusChannel <- "	doorClosedChannel"
			elevator.Direction = orderHandler.Requests_chooseDirection(elevator)
			setMovingDirectionChannel <- elevator.Direction

			updateElevatorInfoChannel <- elevator

		case arrivedAtFloor := <-arrivedAtFloorChannel:
			StatusChannel <- "	arrivedAtFloorChannel"

			// Stop if it should stop
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

		}

	}
}

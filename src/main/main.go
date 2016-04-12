package main

import "driver"
import . "elevator_type"
import "errorHandler"
import "orderHandler"

//import "network"

//import "fmt"

//import . "strconv"

//"129.241.187.156" = workspace 9
//"129.241.187.159" = workspace 11

func main() {

	var elevator ElevatorInfo
	previousFloor := N_FLOORS + 1 // Impossible floor

	setMovingDirectionChannel := make(chan Dir, 1)
	stopChannel := make(chan bool, 1)
	setButtonLightChannel := make(chan ButtonInfo, 1)
	newOrderChannel := make(chan ButtonInfo, 1)
	removeOrderChannel := make(chan ButtonInfo, 1)
	initIsFinished := make(chan bool)
	arrivedAtFloorChannel := make(chan int, 1)
	errorChannel := make(chan string)
	addToRequestsChannel := make(chan ButtonInfo)
	stop := make(chan bool, 1)
	initialElevatorStateChannel := make(chan ElevatorInfo, 1)
	doorClosedChannel := make(chan bool, 1)

	go errorHandler.Error_handler(errorChannel)
	go driver.Driver(setMovingDirectionChannel, stopChannel, setButtonLightChannel, newOrderChannel, initIsFinished, arrivedAtFloorChannel, errorChannel, initialElevatorStateChannel, doorClosedChannel)
	<-initIsFinished
	elevator = <-initialElevatorStateChannel

	go orderHandler.OrderHandler(newOrderChannel, removeOrderChannel, addToRequestsChannel)

	for {
		select {
		case buttonPushed := <-addToRequestsChannel:
			errorChannel <- "1"
			setButtonLightChannel <- buttonPushed
			errorChannel <- "2"
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
						errorChannel <- "In buttonPushed->State_Idle: Direction is zero"
					}
					setMovingDirectionChannel <- Dir(direction)
					elevator.State = State_Moving
				}
			case State_Moving:
				elevator = orderHandler.AddFloorToRequests(elevator, buttonPushed)
			}

		case <-stop:
			errorChannel <- "In stop"
			stopChannel <- true
			elevator.State = State_Idle
			elevator = orderHandler.ClearAtCurrentFloor(elevator)

		case <-doorClosedChannel:
			errorChannel <- "In doorClosedChannel"
			setMovingDirectionChannel <- elevator.Direction
			/////////////////////////
			// UNFINISHED BUSINESS //
			// Direction is never changed to stop
			/////////////////////////

		case arrivedAtFloor := <-arrivedAtFloorChannel:
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
				errorChannel <- "SHOULD STOP"
				stop <- true
			}

		}

	}
}

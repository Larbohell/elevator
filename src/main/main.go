package main

import "elevator"
import . "elevator_type"
import "network"

import . "statusHandler"

import "strconv"

import "flag"
import "os/exec"
import "time"

import "fmt"

func main() {
	var primary bool
	flag.BoolVar(&primary, "primary", true, "Determines if process is primary or backup")
	flag.Parse()

	var startingPoint Message // The information we want to transfer from primary to backup
	errorChannel := make(chan string, 1)
	StatusChannel = make(chan string, 1)

	go Error_handler(errorChannel)
	go Status_handler()

	StatusChannel <- "Backup is here"
	for {
		if primary {
			spawn_backup()

			StatusChannel <- "IN MAIN: Elevator currentFloor = " + strconv.Itoa(startingPoint.ElevatorInfo.CurrentFloor) + " , direction = " + strconv.Itoa(int(startingPoint.ElevatorInfo.Direction)) + ", State: " + strconv.Itoa(int(startingPoint.ElevatorInfo.State))
			/*
				fmt.Printf("\n")

				fmt.Printf("  +--------------------+\n")
				fmt.Printf("  |  | up  | dn  | cab |\n")
				for f := N_FLOORS - 1; f >= 0; f-- {
					fmt.Printf("  | %d", f)
					for btn := 0; btn < N_BUTTONS; btn++ {
						if f == N_FLOORS-1 && btn == int(BUTTON_OUTSIDE_UP) || f == 0 && btn == int(BUTTON_OUTSIDE_DOWN) {
							fmt.Printf("|     ")
						} else {
							if startingPoint.ElevatorInfo.Requests[f][btn] == 1 {
								fmt.Printf("|  #  ")
							} else {
								fmt.Printf("|  -  ")
							}
						}
					}
					fmt.Printf("|\n")
				}
				fmt.Printf("  +--------------------+\n")
			*/
			elevator.Run_elevator(startingPoint, errorChannel)

		} else {
			fmt.Println("Backup process started")
			startingPoint = listenToPrimary()
			primary = true
		}
	}
}

func spawn_backup() {
	run_process_command := exec.Command("gnome-terminal", "-x", "./main", "-primary=0")
	_ = run_process_command.Start()
	//CheckError(err)
}

func listenToPrimary() Message {

	messageFromPrimaryChannel := make(chan Message, 1)
	terminateThreadChannel := make(chan bool, 1)
	threadIsTerminatedChannel := make(chan bool, 1)
	var messageFromPrimary Message

	go network.ReceiveUdpMessageOnOwnIP(messageFromPrimaryChannel, terminateThreadChannel, threadIsTerminatedChannel)
	for {
		select {

		case messageFromPrimary = <-messageFromPrimaryChannel:
			fmt.Println("Primary alive")
			break

		case <-time.After(1 * time.Second):
			fmt.Println("Primary timeout")
			terminateThreadChannel <- true
			fmt.Println("Primary timeout")
			<-threadIsTerminatedChannel
			fmt.Println("Primary timeout")
			return messageFromPrimary

		}
	}
}

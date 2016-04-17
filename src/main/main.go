package main

import "elevator"
import . "elevator_type"

//import "network"

import . "statusHandler"
import "fileHandler"

import "strconv"

import "flag"
import "os/exec"
import "time"

import "fmt"

func main() {
	var primary bool
	flag.BoolVar(&primary, "primary", true, "Determines if process is primary or backup")
	flag.Parse()

	var startingPoint ElevatorInfo // The information we wanerr.Errort to transfer from primary to backup
	var firstTimeRunning bool
	if primary {
		firstTimeRunning = true
	} else {
		firstTimeRunning = false
	}

	errorChannel := make(chan string, 1)
	StatusChannel = make(chan string, 1)

	go Error_handler(errorChannel)
	go Status_handler()

	StatusChannel <- "Backup is here"
	for {
		if primary {
			spawn_backup()

			StatusChannel <- "IN main.go, backup is about to be primary: Elevator currentFloor = " + strconv.Itoa(startingPoint.CurrentFloor) + " , direction = " + strconv.Itoa(int(startingPoint.Direction)) + ", State: " + strconv.Itoa(int(startingPoint.State))
			elevator.Run_elevator(firstTimeRunning, startingPoint, errorChannel)

		} else {
			fmt.Println("Backup process started")
			//startingPoint = listenToPrimary()
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

func listenToPrimary() ElevatorInfo {

	//messageFromPrimaryChannel := make(chan Message, 1)
	terminateThreadChannel := make(chan bool, 1)
	threadIsTerminatedChannel := make(chan bool, 1)
	//var messageFromPrimary Message
	backupFileChangedChannel := make(chan bool, 1)

	//go network.ReceiveUdpMessageOnOwnIP(messageFromPrimaryChannel, terminateThreadChannel, threadIsTerminatedChannel)
	go fileHandler.BackupFileChanged(backupFileChangedChannel, terminateThreadChannel, threadIsTerminatedChannel)
	for {
		select {
		/*
			case messageFromPrimary = <-messageFromPrimaryChannel:
				fmt.Println("Primary alive")
				break
		*/

		case <-backupFileChangedChannel:
			//StatusChannel <- "In backup process: Primary alive, backupFile Changed"

		case <-time.After(1 * time.Second):
			terminateThreadChannel <- true
			<-threadIsTerminatedChannel
			elevator, err := fileHandler.Read()
			if err != nil {
				StatusChannel <- "Error in listenToPrimary: Timeout, error: " + err.Error()
				panic(err)
			}
			StatusChannel <- "In backup timeout (primary dead), CurrentFloor from file read = " + strconv.Itoa(elevator.CurrentFloor)
			return elevator

		}
	}
}

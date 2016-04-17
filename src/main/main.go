package main

import "elevator"
import . "elevator_type"

//import "network"

import . "statusHandler"
import "fileHandler"

//import "strconv"

import "flag"
import "os/exec"
import "time"

import "fmt"

func main() {
	var primary bool
	flag.BoolVar(&primary, "primary", true, "Determines if process is primary or backup")
	flag.Parse()

	var startingPoint ElevatorInfo

	errorChannel := make(chan string, 1)
	StatusChannel = make(chan string, 1)

	go Error_handler(errorChannel)
	go Status_handler()

	for {
		if primary {
			spawn_backup()
			startingPoint, _ = fileHandler.Read()
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
}

func listenToPrimary() ElevatorInfo {

	terminateThreadChannel := make(chan bool, 1)
	threadIsTerminatedChannel := make(chan bool, 1)
	backupFileChangedChannel := make(chan bool, 1)

	go fileHandler.BackupFileChanged(backupFileChangedChannel, terminateThreadChannel, threadIsTerminatedChannel)
	for {
		select {
		case <-backupFileChangedChannel:

		case <-time.After(1 * time.Second):
			terminateThreadChannel <- true
			<-threadIsTerminatedChannel

			elevator, err := fileHandler.Read()
			if err != nil {
				StatusChannel <- "Error in listenToPrimary: Timeout, error: " + err.Error()
				panic(err)
			}

			return elevator
		}
	}
}

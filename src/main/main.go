package main

import "elevator"
import . "elevator_type"
import . "statusHandler"
import "fileHandler"

import "flag"
import "os"
import "os/exec"
import . "time"
import "fmt"

func main() {
	var primary bool
	flag.BoolVar(&primary, "primary", true, "Determines if process is primary or backup")
	flag.Parse()

	var startingPoint ElevatorInfo

	errorChannel := make(chan string, 1)
	StatusChannel = make(chan string, 1)
	programAliveChannel := make(chan bool, 1)

	go Error_handler(errorChannel)
	go Status_handler()

	for {
		if primary {
			go exitProgramIfTimeout(programAliveChannel)
			spawn_backup()
			startingPoint, _ = fileHandler.Read()
			elevator.Run_elevator(startingPoint, errorChannel, programAliveChannel)

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

		case <-After(1 * Second):
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

func exitProgramIfTimeout(programAliveChannel chan bool) {
	for {
		select {
		case <-programAliveChannel:
			break
		case <-After(5 * Second):
			os.Exit(1)
		}
	}
}

package fileHandler

import . "elevator_type"
import "time"
import "os"
import "fmt"

//import "strconv"

import . "statusHandler"

const BACKUP_FILE_NAME string = "backup.txt"

//func Write(numOfElevs int, numOfFloors int, length int, queue []int) {
func write(elevator ElevatorInfo) {
	//StatusChannel <- "Writing to backup file"
	file, err := os.Create(BACKUP_FILE_NAME)
	if err != nil {
		StatusChannel <- "Error creating file '" + BACKUP_FILE_NAME + "'"
		panic(err)
	}

	defer func() {
		err = file.Close()
		if err != nil {
			StatusChannel <- "Error writing to backup file"
			panic(err)
		}
	}()

	_, err = fmt.Fprintf(file, "%d\n", elevator.CurrentFloor)
	if err != nil {
		panic(err)
	}

	_, err = fmt.Fprintf(file, "%d\n", int(elevator.Direction))
	if err != nil {
		panic(err)
	}

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			_, err = fmt.Fprintf(file, "%d\n", elevator.Requests[floor][btn])
			if err != nil {
				panic(err)
			}
		}
	}

	_, err = fmt.Fprintf(file, "%d\n", int(elevator.State))
	if err != nil {
		panic(err)
	}
}

func Read() (ElevatorInfo, error) {

	var elevator ElevatorInfo

	file, err := os.Open(BACKUP_FILE_NAME)
	if err != nil {
		//StatusChannel <- err
		return elevator, err
	}

	defer func() {
		err = file.Close()
		if err != nil {
			StatusChannel <- "Error reading from backup file"
			panic(err)
		}
	}()

	var buf int

	_, err = fmt.Fscanf(file, "%d\n", &buf)
	if err != nil {
		panic(err)
	} else {
		elevator.CurrentFloor = buf
	}

	_, err = fmt.Fscanf(file, "%d\n", &buf)
	if err != nil {
		panic(err)
	} else {
		elevator.Direction = Dir(buf)
	}

	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			_, err = fmt.Fscanf(file, "%d\n", &buf)
			elevator.Requests[floor][btn] = buf
			if err != nil {
				panic(err)
			}
		}
	}

	_, err = fmt.Fscanf(file, "%d\n", &buf)
	if err != nil {
		panic(err)
	} else {
		elevator.State = ElevatorState(buf)
	}

	return elevator, err
}

func BackupElevatorInfoToFile(backupChannel chan ElevatorInfo) {
	var elevator ElevatorInfo

	for {
		select {
		case elevator = <-backupChannel:
			write(elevator)

		case <-time.After(50 * time.Millisecond):
			write(elevator)
		}
	}
}

func BackupFileChanged(backupFileChangedChannel chan bool, terminateThreadChannel chan bool, threadIsTerminatedChannel chan bool) {
	var lastModifiedTime time.Time

	for {
		time.Sleep(50 * time.Millisecond)

		info, err := os.Stat(BACKUP_FILE_NAME)
		if err != nil {
			StatusChannel <- "Error in backupFileChanged: " + err.Error()
		}

		modifiedTime := info.ModTime()
		if modifiedTime != lastModifiedTime {
			backupFileChangedChannel <- true
		}
		lastModifiedTime = modifiedTime

		select {
		case <-terminateThreadChannel:
			threadIsTerminatedChannel <- true
			return
		default:
			break
		}
	}

}

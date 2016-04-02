package elevator

import "time"
import "fmt"

//var timerEndTime time.Time
//var timerActive bool

func Timer_start(){
	
	/*fmt.Printf("Timer started\n")
	timerStartTime := time.Now()
	timerEndTime = timerStartTime.Add(3 * time.Second)
	timerActive = true
	*/
	
	fmt.Printf("Timer started\n")
	duration := time.Duration(3 * time.Second)
	time.Sleep(duration)
	fmt.Printf("Timer ended\n")
	Fsm_onDoorTimeout()
}
/*
func Timer_timedout() bool {
	currentTime := time.Now()
	return timerActive && currentTime.After(timerEndTime)
}

func Timer_stop() {
	timerActive = false
}
*/
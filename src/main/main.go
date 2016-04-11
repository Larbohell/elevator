package main

import "driver"
import "elevator_type"
//import "network"

//import "fmt"
//import "strconv"

//"129.241.187.156" = workspace 9
//"129.241.187.159" = workspace 11

func main() {

		setMovingDirectionChannel := make(chan elevator_type.Dir, 1)
		stopChannel := make(chan bool, 1)
		setButtonLightChannel := make(chan elevator_type.ButtonInfo, 1)
		newOrderChannel := make(chan elevator_type.ButtonInfo, 1)
		initIsFinished := make(chan bool)


		go driver.Driver(setMovingDirectionChannel, stopChannel, setButtonLightChannel, newOrderChannel, initIsFinished)
}
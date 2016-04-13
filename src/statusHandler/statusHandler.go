package statusHandler

import . "fmt"

var StatusChannel chan string

func Status_handler() {

	for {
		statusMessage := <-StatusChannel
		Println("\x1b[32;1m" + statusMessage + "\x1b[0m" + "\n")
	}
}

func Error_handler(errorChannel chan string) {
	for {
		errorMsg := <-errorChannel
		Println("\x1b[31;1m" + errorMsg + "\x1b[0m" + "\n")
	}
}
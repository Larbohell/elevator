package errorHandler

import "fmt"

func Error_handler(errorChannel chan string) {
	for {
		errorMsg := <-errorChannel
		fmt.Println(errorMsg)
	}
}

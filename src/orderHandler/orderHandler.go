package orderHandler

import . "elevator_type"
import . "statusHandler"

func OrderHandler(addOrderChannel chan ButtonInfo, removeOrderChannel chan ButtonInfo, addToRequestsChannel chan ButtonInfo, externalOrderChannel chan ButtonInfo) {
	StatusChannel <- "In OrderHandler"

	for {
		select {
		case addOrder := <-addOrderChannel:
			if addOrder.Button == BUTTON_INSIDE_COMMAND {
				StatusChannel <- "	Button pushed = INSIDE_COMMAND"
				addToRequestsChannel <- addOrder
			} else {
				// Inform master, master handles order and adds to one of the elevator's queues
				StatusChannel <- "	Button pushed = EXTERNAL"
				externalOrderChannel <- addOrder
			}
		}
	}
}

func BestElevatorForTheJob(button ButtonInfo) string {
	return "129.24.187.159"
}

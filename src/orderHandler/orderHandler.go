package orderHandler

import . "elevator_type"

func OrderHandler(addOrderChannel chan ButtonInfo, removeOrderChannel chan ButtonInfo, addToRequestsChannel chan ButtonInfo, externalOrderChannel chan ButtonInfo) {

	for {
		select {
		case addOrder := <-addOrderChannel:
			if addOrder.Button == BUTTON_INSIDE_COMMAND {
				addToRequestsChannel <- addOrder
			} else {
				// Inform master, master handles order and adds to one of the elevator's queues
				externalOrderChannel <- addOrder
			}
		}
	}
}
